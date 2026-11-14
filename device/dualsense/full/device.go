package dualsensefull

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"net"
	"slices"
	"strconv"
	"sync"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/dualsense"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/Alia5/VIIPER/viiperclient"
)

type CreateOptions struct {
	Meta      dualsense.MetaState
	AudioPort *uint16 `json:"audio_port,omitempty"`
}

type audioConn struct {
	conn net.Conn
	out  chan []byte
	done chan struct{}
}

type audioFeatureUnit struct {
	mute   bool
	volume int16
}

type Device struct {
	ctx context.Context

	*dualsense.DualSense

	descriptor usb.Descriptor
	edge       bool

	audioListener net.Listener
	audioConns    []audioConn
	audioMtx      sync.Mutex

	microphone            []byte
	speakerFeatureUnit    audioFeatureUnit
	microphoneFeatureUnit audioFeatureUnit
	altSettings           [interfaceCount]uint8

	outputFunc      func(OutputState)
	microphoneMuted bool
}

func New(ctx context.Context, o *device.CreateOptions) (*Device, error) {
	return new(ctx, o, dualsense.New, false)
}

func NewEdge(ctx context.Context, o *device.CreateOptions) (*Device, error) {
	return new(ctx, o, dualsense.NewEdge, true)
}

func new(ctx context.Context, o *device.CreateOptions, newParent func(context.Context, *device.CreateOptions) (*dualsense.DualSense, error), edge bool) (*Device, error) {
	var audioListener net.Listener
	if o != nil && o.DeviceSpecific != "" {
		var createOptions CreateOptions
		err := json.Unmarshal([]byte(o.DeviceSpecific), &createOptions)
		if err != nil {
			return nil, viiperclient.ErrBadRequest(fmt.Sprintf("invalid device specific JSON: %v", err))
		}
		meta, err := json.Marshal(createOptions.Meta)
		if err != nil {
			return nil, viiperclient.ErrInternal(fmt.Sprintf("marshal meta state: %v", err))
		}
		o.DeviceSpecific = string(meta)
		if createOptions.AudioPort != nil {
			if *createOptions.AudioPort == 0 {
				return nil, viiperclient.ErrBadRequest("invalid audio_port: 0")
			}
			conn := device.GetConn(ctx)
			if conn == nil {
				audioListener, err = net.Listen(
					"tcp",
					net.JoinHostPort(
						"localhost",
						strconv.Itoa(int(*createOptions.AudioPort)),
					),
				)
			} else {
				local := conn.LocalAddr().(*net.TCPAddr)
				audioListener, err = net.ListenTCP(
					"tcp",
					&net.TCPAddr{
						IP:   local.IP,
						Port: int(*createOptions.AudioPort),
					},
				)
			}
			if err != nil {
				return nil, viiperclient.ErrConflict(fmt.Sprintf("audio_port %d: %v", *createOptions.AudioPort, err))
			}
		}
	}
	parent, err := newParent(ctx, o)
	if err != nil {
		if audioListener != nil {
			_ = audioListener.Close()
		}
		return nil, err
	}
	d := &Device{
		ctx:           ctx,
		DualSense:     parent,
		descriptor:    defaultDescriptor,
		edge:          edge,
		audioListener: audioListener,
		speakerFeatureUnit: audioFeatureUnit{
			volume: speakerVolumeMax,
		},
		microphoneFeatureUnit: audioFeatureUnit{
			volume: microphoneVolumeMax,
		},
	}
	d.descriptor.Device.IDProduct = dualsense.DefaultPIDDS
	if edge {
		d.descriptor.Device.IDProduct = dualsense.DefaultPIDDSEdge
		d.descriptor.Strings = maps.Clone(d.descriptor.Strings)
		d.descriptor.Strings[2] = "DualSense Edge Wireless Controller"

		hidFunction := *d.descriptor.Interfaces[5].HID
		gamepadCollection := hidFunction.ReportDescriptor.Items[2].(hid.Collection)
		gamepadCollection.Items = slices.Clone(gamepadCollection.Items)
		for i, item := range gamepadCollection.Items {
			reportID, isReportID := item.(hid.ReportID)
			if !isReportID {
				continue
			}
			switch reportID.ID {
			case dualsense.ReportIDOutput:
				gamepadCollection.Items[i+dualsense.ReportCountOffsetFromReportID] = hid.ReportCount{
					Count: dualsense.OutputReportCountEdge,
				}
			case 0xF2:
				gamepadCollection.Items[i+dualsense.ReportCountOffsetFromReportID] = hid.ReportCount{
					Count: dualsense.FeatureF2CountEdge,
				}
			}
		}
		gamepadCollection.Items = append(gamepadCollection.Items, dualsense.EdgeOnlyFeatureReports...)
		hidFunction.ReportDescriptor.Items = slices.Clone(hidFunction.ReportDescriptor.Items)
		hidFunction.ReportDescriptor.Items[2] = gamepadCollection
		d.descriptor.Interfaces = slices.Clone(d.descriptor.Interfaces)
		d.descriptor.Interfaces[5].HID = &hidFunction
	}

	if o != nil {
		if o.IDVendor != nil {
			d.descriptor.Device.IDVendor = *o.IDVendor
		}
		if o.IDProduct != nil {
			d.descriptor.Device.IDProduct = *o.IDProduct
		}
	}

	if d.audioListener != nil {
		go d.acceptAudio()
	}

	return d, nil
}

func (d *Device) acceptAudio() {
	go func() {
		<-d.ctx.Done()
		_ = d.audioListener.Close()
		d.audioMtx.Lock()
		for _, c := range d.audioConns {
			_ = c.conn.Close()
		}
		d.audioMtx.Unlock()
	}()
	for {
		conn, err := d.audioListener.Accept()
		if err != nil {
			return
		}
		c := audioConn{
			conn: conn,
			out:  make(chan []byte, audioOutQueueSize),
			done: make(chan struct{}),
		}
		d.audioMtx.Lock()
		d.audioConns = append(d.audioConns, c)
		d.audioMtx.Unlock()
		go d.writeAudio(c)
		go d.readAudio(c)
	}
}

func (d *Device) writeAudio(c audioConn) {
	for {
		select {
		case <-c.done:
			return
		case pcm := <-c.out:
			_, err := c.conn.Write(pcm)
			if err != nil {
				_ = c.conn.Close()
				return
			}
		}
	}
}

func (d *Device) readAudio(c audioConn) {
	buf := make([]byte, audioReadBufferSize)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			break
		}
		d.audioMtx.Lock()
		d.microphone = append(d.microphone, buf[:n]...)
		excess := len(d.microphone) - microphoneBufferSize
		if excess > 0 {
			excess += microphoneFrameSize - 1
			excess -= excess % microphoneFrameSize
			d.microphone = d.microphone[excess:]
		}
		d.audioMtx.Unlock()
	}
	_ = c.conn.Close()
	close(c.done)
	d.audioMtx.Lock()
	d.audioConns = slices.DeleteFunc(d.audioConns, func(other audioConn) bool {
		return other.conn == c.conn
	})
	d.audioMtx.Unlock()
}

func (d *Device) DeviceTypeName() string {
	if d.edge {
		return TypeNameEdge.Name
	}
	return TypeName.Name
}

func (d *Device) GetDescriptor() *usb.Descriptor {
	return &d.descriptor
}

func (d *Device) SetOutputCallback(f func(OutputState)) {
	d.audioMtx.Lock()
	d.outputFunc = f
	d.audioMtx.Unlock()
}

func (d *Device) handleOutputReport(report []byte) {
	if len(report) <= dualsense.OutputReportOffsetLedBlue || report[0] != dualsense.ReportIDOutput {
		return
	}
	state := OutputState{
		OutputState:      dualsense.ParseOutputReport(report),
		SpeakerVolume:    report[OutputReportOffsetSpeakerVolume],
		MicrophoneVolume: report[OutputReportOffsetMicrophoneVolume],
		AudioControl:     report[OutputReportOffsetAudioControl],
		MicMuteLed:       report[OutputReportOffsetMicMuteLed],
		PowerSaveControl: report[OutputReportOffsetPowerSaveControl],
		AudioControl2:    report[OutputReportOffsetAudioControl2],
	}
	d.audioMtx.Lock()
	if state.UpdateFlags2&dualsense.UpdateFlag2PowerSaveControl != 0 {
		d.microphoneMuted = state.PowerSaveControl&PowerSaveControlMicrophoneMute != 0
	}
	outputFunc := d.outputFunc
	d.audioMtx.Unlock()
	if outputFunc != nil {
		outputFunc(state)
	}
}

func (d *Device) setMicrophoneMuteState(report []byte) {
	if len(report) != dualsense.InputReportSize || report[0] != dualsense.ReportIDInput {
		return
	}
	d.audioMtx.Lock()
	defer d.audioMtx.Unlock()
	if d.microphoneMuted {
		report[dualsense.InputReportOffsetConnectState] |= ConnectStateMicrophoneMuted
	}
}

func (d *Device) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) ([]byte, bool) {
	if bmRequestType == 0x01 && bRequest == requestSetInterface && wIndex < interfaceCount {
		d.audioMtx.Lock()
		d.altSettings[wIndex] = uint8(wValue)
		d.audioMtx.Unlock()
		return nil, true
	}
	if bmRequestType == 0x81 && bRequest == requestGetInterface && wIndex < interfaceCount {
		d.audioMtx.Lock()
		alt := d.altSettings[wIndex]
		d.audioMtx.Unlock()
		return []byte{alt}, true
	}
	if uint8(wIndex) == interfaceAudioControl && (bmRequestType == audioRequestTypeSet || bmRequestType == audioRequestTypeGet) {
		return d.handleAudioControl(bmRequestType, audioRequest(bRequest), wValue, wIndex, data)
	}

	if bmRequestType == dualsense.HIDClassOUT && bRequest == dualsense.HIDSetReport && uint8(wValue>>8) == dualsense.ReportTypeOutput {
		d.handleOutputReport(data)
	}
	report, handled := d.DualSense.HandleControl(bmRequestType, bRequest, wValue, wIndex, wLength, data)
	d.setMicrophoneMuteState(report)
	return report, handled
}

func (d *Device) handleAudioControl(bmRequestType uint8, request audioRequest, wValue, wIndex uint16, data []byte) ([]byte, bool) {
	var featureUnit *audioFeatureUnit
	var minimum, maximum, resolution int16
	switch uint8(wIndex >> 8) {
	case audioEntitySpeakerFeatureUnit:
		featureUnit = &d.speakerFeatureUnit
		minimum = speakerVolumeMin
		maximum = speakerVolumeMax
		resolution = speakerVolumeRes
	case audioEntityMicrophoneFeatureUnit:
		featureUnit = &d.microphoneFeatureUnit
		minimum = microphoneVolumeMin
		maximum = microphoneVolumeMax
		resolution = microphoneVolumeRes
	default:
		return nil, false
	}

	d.audioMtx.Lock()
	defer d.audioMtx.Unlock()

	selector := uint8(wValue >> 8)
	if bmRequestType == audioRequestTypeSet {
		if request != audioRequestSetCur {
			return nil, false
		}
		if selector == audioControlMute && len(data) == 1 {
			featureUnit.mute = data[0] != 0
			return nil, true
		}
		if selector == audioControlVolume && len(data) == 2 {
			featureUnit.volume = int16(binary.LittleEndian.Uint16(data))
			return nil, true
		}
		return nil, false
	}

	if selector == audioControlMute && request == audioRequestGetCur {
		if featureUnit.mute {
			return []byte{1}, true
		}
		return []byte{0}, true
	}
	if selector != audioControlVolume {
		return nil, false
	}
	value := featureUnit.volume
	switch request {
	case audioRequestGetCur:
	case audioRequestGetMin:
		value = minimum
	case audioRequestGetMax:
		value = maximum
	case audioRequestGetRes:
		value = resolution
	default:
		return nil, false
	}
	return binary.LittleEndian.AppendUint16(nil, uint16(value)), true
}

func (d *Device) HandleTransfer(ctx context.Context, ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirOut && ep == endpointAudioOut {
		pcm := append([]byte(nil), out...)
		d.audioMtx.Lock()
		applyVolume(pcm, d.speakerFeatureUnit, speakerVolumeMax)
		for _, c := range d.audioConns {
			select {
			case c.out <- pcm:
			default:
			}
		}
		d.audioMtx.Unlock()
		return nil
	}
	if dir == usbip.DirIn && ep == endpointMicrophoneIn {
		pcm := make([]byte, AudioMicrophonePacketSize)
		d.audioMtx.Lock()
		if len(d.microphone) >= AudioMicrophonePacketSize {
			copy(pcm, d.microphone)
			d.microphone = d.microphone[AudioMicrophonePacketSize:]
		}
		applyVolume(pcm, d.microphoneFeatureUnit, microphoneVolumeMax)
		d.audioMtx.Unlock()
		return pcm
	}

	if dir == usbip.DirOut && ep == dualsense.EndpointOut {
		d.handleOutputReport(out)
	}
	report := d.DualSense.HandleTransfer(ctx, ep, dir, out)
	d.setMicrophoneMuteState(report)
	return report
}

func applyVolume(pcm []byte, featureUnit audioFeatureUnit, maximum int16) {
	if featureUnit.mute || featureUnit.volume == audioVolumeSilence {
		clear(pcm)
		return
	}
	if featureUnit.volume >= maximum {
		return
	}
	decibel := float64(int(featureUnit.volume)-int(maximum)) / float64(audioVolumeDecibel)
	gain := math.Pow(10, decibel/20)
	for i := 0; i+AudioBytesPerSample <= len(pcm); i += AudioBytesPerSample {
		sample := int16(binary.LittleEndian.Uint16(pcm[i:]))
		binary.LittleEndian.PutUint16(pcm[i:], uint16(int16(float64(sample)*gain)))
	}
}
