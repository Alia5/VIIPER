package main

/*

Example of accessing a DualSense/Edge via SDL and passing the real device through to a VIIPER emulated DualSense/Edge via VIIPER TCP Api
Run VIIPER, then this forwarder, and connect a real DualSense controller.

You can use [HidHide](https://docs.nefarius.at/projects/HidHide/) to hide the real controller from other apps.
Needs SDL3.dll (built from main branch, newer than 7b82a40bbf1aad01d9b0f1f3bc18273a1ef9b6b4) and cgo.
*/

import (
	"bufio"
	"context"
	"encoding"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/Alia5/VIIPER/device/dualsense"
	dualsensefull "github.com/Alia5/VIIPER/device/dualsense/full"
	"github.com/Alia5/VIIPER/examples/go/dualsense_passthrough/sdl"
	"github.com/Alia5/VIIPER/viiperclient"
)

const (
	eventWaitTimeout = 16 * time.Millisecond
	degreesPerRadian = 180 / math.Pi

	audioPort uint16 = 5000
)

var sdlHints = map[sdl.Hint]string{
	sdl.HintJoystickAllowBackgroundEvents: "1",
	sdl.HintJoystickHIDAPI:                "1",
	sdl.HintJoystickHIDAPIPS5:             "1",
	sdl.HintJoystickEnhancedReports:       "1",
}

func main() {
	addr := flag.String("addr", "localhost:3242", "VIIPER API address")
	flag.Parse()

	for hint, value := range sdlHints {
		err := sdl.SetHint(hint, value)
		if err != nil {
			slog.Error("Setting an SDL hint failed", "hint", hint, "error", err)
			return
		}
	}
	err := sdl.InitSubSystem(sdl.InitFlagGamepad | sdl.InitFlagSensor | sdl.InitFlagAudio)
	if err != nil {
		slog.Error("Initializing SDL failed", "error", err)
		return
	}
	defer sdl.Quit()

	slog.Info("Waiting for a DualSense controller")
	gamepad, err := sdl.OpenGamepad(waitForDualSense())
	if err != nil {
		slog.Error("Opening the controller failed", "error", err)
		return
	}
	defer gamepad.Close()
	if !gamepad.SetSensorEnabled(sdl.SensorTypeGyroscope, true) {
		slog.Warn("Enabling the gyroscope failed", "error", sdl.GetError())
	}
	if !gamepad.SetSensorEnabled(sdl.SensorTypeAccelerometer, true) {
		slog.Warn("Enabling the accelerometer failed", "error", sdl.GetError())
	}
	slog.Info(
		"Opened the controller",
		"name", gamepad.Name(),
		"serial", gamepad.Serial(),
		"path", gamepad.Path(),
	)

	audioHost, _, err := net.SplitHostPort(*addr)
	if err != nil {
		slog.Error("The VIIPER API address is invalid", "addr", *addr, "error", err)
		return
	}
	client := viiperclient.New(*addr)
	buses, err := client.BusList()
	if err != nil {
		slog.Error("Listing the VIIPER buses failed", "error", err)
		return
	}
	var busID uint32
	if len(buses.Buses) == 0 {
		bus, err := client.BusCreate(0)
		if err != nil {
			slog.Error("Creating a VIIPER bus failed", "error", err)
			return
		}
		busID = bus.BusID
		defer func() {
			_, _ = client.BusRemove(busID)
		}()
	} else {
		busID = slices.Min(buses.Buses)
	}

	deviceType := dualsensefull.TypeNameEdge
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, device, err := client.AddDeviceAndConnect(
		ctx,
		busID,
		deviceType,
		&viiperclient.CreateOptions[dualsensefull.CreateOptions]{
			DeviceSpecific: &dualsensefull.CreateOptions{
				Meta: dualsense.MetaState{
					ShellColor: dualsense.ShellColorAstroBotJoyfulEdition,
				},
				AudioPort: new(audioPort),
			},
		},
	)
	if err != nil {
		slog.Error("Creating the VIIPER device failed", "type", deviceType.Name, "error", err)
		return
	}
	defer func() {
		_ = stream.Close()
		_, _ = client.DeviceRemove(device.BusID, device.DevID)
		gamepad.Rumble(0, 0, 0)
	}()
	slog.Info("Created the VIIPER device", "type", deviceType.Name, "bus", device.BusID, "dev", device.DevID)

	feedback, feedbackErr := stream.StartReading(
		ctx,
		32,
		func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
			var b [dualsensefull.OutputStateSize]byte
			_, err := io.ReadFull(r, b[:])
			if err != nil {
				return nil, err
			}
			output := new(dualsensefull.OutputState)
			return output, output.UnmarshalBinary(b[:])
		},
	)
	go forwardFeedback(ctx, cancel, gamepad, feedback, feedbackErr)

	err = forwardAudio(ctx, gamepad, audioHost)
	if err != nil {
		slog.Error("Starting the audio passthrough failed", "error", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	state := dualsense.NewInputState()
	for {
		select {
		case <-interrupt:
			return
		case <-ctx.Done():
			return
		default:
		}

		ev, _ := sdl.WaitEventTimeout(eventWaitTimeout)
		switch ev := ev.(type) {
		case *sdl.QuitEvent:
			return
		case *sdl.GamepadSensorEvent:
			switch sdl.SensorType(ev.Sensor) {
			case sdl.SensorTypeGyroscope:
				state.GyroX = dualsense.GyroDpsToRaw(float64(ev.Data0) * degreesPerRadian)
				state.GyroY = dualsense.GyroDpsToRaw(float64(ev.Data1) * degreesPerRadian)
				state.GyroZ = dualsense.GyroDpsToRaw(float64(ev.Data2) * degreesPerRadian)
			case sdl.SensorTypeAccelerometer:
				state.AccelX = dualsense.AccelMS2ToRaw(float64(ev.Data0))
				state.AccelY = dualsense.AccelMS2ToRaw(float64(ev.Data1))
				state.AccelZ = dualsense.AccelMS2ToRaw(float64(ev.Data2))
			}
		case *sdl.GamepadTouchpadEvent:
			active := ev.Type != sdl.EventTypeGamepadTouchpadUp
			switch ev.Finger {
			case 0:
				state.Touch1Active = active
				if active {
					state.Touch1X = uint16(math.Round(float64(ev.X) * float64(dualsense.TouchpadMaxX)))
					state.Touch1Y = uint16(math.Round(float64(ev.Y) * float64(dualsense.TouchpadMaxY)))
				}
			case 1:
				state.Touch2Active = active
				if active {
					state.Touch2X = uint16(math.Round(float64(ev.X) * float64(dualsense.TouchpadMaxX)))
					state.Touch2Y = uint16(math.Round(float64(ev.Y) * float64(dualsense.TouchpadMaxY)))
				}
			}
		case *sdl.GamepadDeviceEvent:
			switch ev.Type {
			case sdl.EventTypeGamepadUpdateComplete:
				err := sendState(gamepad, state, stream)
				if err != nil {
					slog.Error("Sending the input state to the VIIPER device failed", "error", err)
					return
				}
			case sdl.EventTypeGamepadRemoved:
				if sdl.GamepadID(ev.Which) == gamepad.ID() {
					slog.Info("The controller was disconnected")
					return
				}
			}
		}
	}
}

func waitForDualSense() sdl.GamepadID {
	for {
		ev, _ := sdl.WaitEventTimeout(eventWaitTimeout)
		added, ok := ev.(*sdl.GamepadDeviceEvent)
		if !ok || added.Type != sdl.EventTypeGamepadAdded {
			continue
		}
		id := sdl.GamepadID(added.Which)
		if sdl.GetGamepadVendorForID(id) != dualsense.DefaultVID {
			continue
		}
		product := sdl.GetGamepadProductForID(id)
		if product == dualsense.DefaultPIDDS || product == dualsense.DefaultPIDDSEdge {
			return id
		}
	}
}

func sendState(gp *sdl.Gamepad, s *dualsense.InputState, stream *viiperclient.DeviceStream) error {
	s.Buttons = 0
	if gp.GetButton(sdl.GamepadButtonSouth) {
		s.Buttons |= dualsense.ButtonCross
	}
	if gp.GetButton(sdl.GamepadButtonEast) {
		s.Buttons |= dualsense.ButtonCircle
	}
	if gp.GetButton(sdl.GamepadButtonWest) {
		s.Buttons |= dualsense.ButtonSquare
	}
	if gp.GetButton(sdl.GamepadButtonNorth) {
		s.Buttons |= dualsense.ButtonTriangle
	}
	if gp.GetButton(sdl.GamepadButtonLeftShoulder) {
		s.Buttons |= dualsense.ButtonL1
	}
	if gp.GetButton(sdl.GamepadButtonRightShoulder) {
		s.Buttons |= dualsense.ButtonR1
	}
	if gp.GetButton(sdl.GamepadButtonBack) {
		s.Buttons |= dualsense.ButtonCreate
	}
	if gp.GetButton(sdl.GamepadButtonStart) {
		s.Buttons |= dualsense.ButtonOptions
	}
	if gp.GetButton(sdl.GamepadButtonLeftStick) {
		s.Buttons |= dualsense.ButtonL3
	}
	if gp.GetButton(sdl.GamepadButtonRightStick) {
		s.Buttons |= dualsense.ButtonR3
	}
	if gp.GetButton(sdl.GamepadButtonGuide) {
		s.Buttons |= dualsense.ButtonPS
	}
	if gp.GetButton(sdl.GamepadButtonTouchpad) {
		s.Buttons |= dualsense.ButtonTouchpad
	}
	if gp.GetButton(sdl.GamepadButtonMisc1) {
		s.Buttons |= dualsense.ButtonMicMute
	}
	if gp.GetButton(sdl.GamepadButtonLeftPaddle1) {
		s.Buttons |= dualsense.ButtonEdgeL4
	}
	if gp.GetButton(sdl.GamepadButtonRightPaddle1) {
		s.Buttons |= dualsense.ButtonEdgeR4
	}
	if gp.GetButton(sdl.GamepadButtonLeftPaddle2) {
		s.Buttons |= dualsense.ButtonEdgeLFn
	}
	if gp.GetButton(sdl.GamepadButtonRightPaddle2) {
		s.Buttons |= dualsense.ButtonEdgeRFn
	}

	s.DPad = 0
	if gp.GetButton(sdl.GamepadButtonDpadUp) {
		s.DPad |= dualsense.DPadUp
	}
	if gp.GetButton(sdl.GamepadButtonDpadDown) {
		s.DPad |= dualsense.DPadDown
	}
	if gp.GetButton(sdl.GamepadButtonDpadLeft) {
		s.DPad |= dualsense.DPadLeft
	}
	if gp.GetButton(sdl.GamepadButtonDpadRight) {
		s.DPad |= dualsense.DPadRight
	}

	s.LX = int8(int(gp.GetAxis(sdl.GamepadAxisLeftX)) * math.MaxInt8 / math.MaxInt16)
	s.LY = int8(int(gp.GetAxis(sdl.GamepadAxisLeftY)) * math.MaxInt8 / math.MaxInt16)
	s.RX = int8(int(gp.GetAxis(sdl.GamepadAxisRightX)) * math.MaxInt8 / math.MaxInt16)
	s.RY = int8(int(gp.GetAxis(sdl.GamepadAxisRightY)) * math.MaxInt8 / math.MaxInt16)
	s.L2 = uint8(int(gp.GetAxis(sdl.GamepadAxisLeftTrigger)) * math.MaxUint8 / math.MaxInt16)
	s.R2 = uint8(int(gp.GetAxis(sdl.GamepadAxisRightTrigger)) * math.MaxUint8 / math.MaxInt16)
	if s.L2 > 0 {
		s.Buttons |= dualsense.ButtonL2
	}
	if s.R2 > 0 {
		s.Buttons |= dualsense.ButtonR2
	}

	return stream.WriteBinary(s)
}

func forwardFeedback(
	ctx context.Context,
	cancel context.CancelFunc,
	gamepad *sdl.Gamepad,
	feedback <-chan encoding.BinaryUnmarshaler,
	feedbackErr <-chan error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-feedbackErr:
			slog.Error("Reading the feedback of the VIIPER device failed", "error", err)
			cancel()
			return
		case message := <-feedback:
			f := message.(*dualsensefull.OutputState)
			feedbackLog := []any{
				"flags", fmt.Sprintf("%02x/%02x/%02x", f.UpdateFlags1, f.UpdateFlags2, f.UpdateFlags3),
			}
			if f.UpdateFlags1&(dualsense.UpdateFlag1Rumble|dualsense.UpdateFlag1HapticsSelect) != 0 {
				feedbackLog = append(feedbackLog, "rumbleSmall", f.RumbleSmall, "rumbleLarge", f.RumbleLarge)
			}
			if f.UpdateFlags1&dualsense.UpdateFlag1TriggerR2 != 0 {
				feedbackLog = append(feedbackLog, "triggerR2", fmt.Sprintf("%+v", f.TriggerR2()))
			}
			if f.UpdateFlags1&dualsense.UpdateFlag1TriggerL2 != 0 {
				feedbackLog = append(feedbackLog, "triggerL2", fmt.Sprintf("%+v", f.TriggerL2()))
			}
			if f.UpdateFlags1&dualsense.UpdateFlag1SpeakerVolume != 0 {
				feedbackLog = append(feedbackLog, "speakerVolume", f.SpeakerVolume)
			}
			if f.UpdateFlags1&dualsense.UpdateFlag1MicrophoneVolume != 0 {
				feedbackLog = append(feedbackLog, "microphoneVolume", f.MicrophoneVolume)
			}
			if f.UpdateFlags1&dualsense.UpdateFlag1AudioControl != 0 {
				feedbackLog = append(feedbackLog, "audioControl", fmt.Sprintf("%02x", f.AudioControl))
			}
			if f.UpdateFlags2&dualsense.UpdateFlag2MicMuteLed != 0 {
				feedbackLog = append(feedbackLog, "micMuteLed", f.MicMuteLed)
			}
			if f.UpdateFlags2&dualsense.UpdateFlag2PowerSaveControl != 0 {
				feedbackLog = append(feedbackLog, "powerSaveControl", fmt.Sprintf("%02x", f.PowerSaveControl))
			}
			if f.UpdateFlags2&dualsense.UpdateFlag2AudioControl2 != 0 {
				feedbackLog = append(feedbackLog, "audioControl2", fmt.Sprintf("%02x", f.AudioControl2))
			}
			if f.UpdateFlags2&dualsense.UpdateFlag2Led != 0 {
				feedbackLog = append(feedbackLog, "led", fmt.Sprintf("%02x%02x%02x", f.LedRed, f.LedGreen, f.LedBlue))
			}
			if f.UpdateFlags2&dualsense.UpdateFlag2PlayerLeds != 0 {
				feedbackLog = append(feedbackLog, "playerLeds", fmt.Sprintf("%06b", f.PlayerLeds))
			}
			slog.Info("Received feedback from the VIIPER device", feedbackLog...)

			rumble := f.UpdateFlags1 & (dualsense.UpdateFlag1Rumble | dualsense.UpdateFlag1HapticsSelect)
			triggers := f.UpdateFlags1 & (dualsense.UpdateFlag1TriggerR2 | dualsense.UpdateFlag1TriggerL2)
			playerLeds := f.UpdateFlags2 & dualsense.UpdateFlag2PlayerLeds
			microphone := f.UpdateFlags2 & (dualsense.UpdateFlag2MicMuteLed |
				dualsense.UpdateFlag2PowerSaveControl)
			speaker := f.UpdateFlags1 & (dualsense.UpdateFlag1SpeakerVolume |
				dualsense.UpdateFlag1AudioControl)
			speakerGain := f.UpdateFlags2 & dualsense.UpdateFlag2AudioControl2

			var rumbleLarge, rumbleSmall uint16
			if rumble != 0 {
				rumbleLarge = uint16(f.RumbleLarge) * (math.MaxUint16 / math.MaxUint8)
				rumbleSmall = uint16(f.RumbleSmall) * (math.MaxUint16 / math.MaxUint8)
			}
			if !gamepad.Rumble(rumbleLarge, rumbleSmall, math.MaxUint32) {
				slog.Error("Setting the rumble of the controller failed", "error", sdl.GetError())
			}

			if f.UpdateFlags2&dualsense.UpdateFlag2Led != 0 {
				if !gamepad.SetLED(f.LedRed, f.LedGreen, f.LedBlue) {
					slog.Error("Setting the LED color of the controller failed", "error", sdl.GetError())
				}
			}

			if triggers == 0 && playerLeds == 0 && microphone == 0 && speaker == 0 && speakerGain == 0 {
				continue
			}
			effects := *f
			effects.UpdateFlags1 = rumble | triggers | speaker
			effects.UpdateFlags2 = playerLeds | microphone | speakerGain
			effects.UpdateFlags3 &= dualsense.UpdateFlag3ImprovedRumble
			// The other AudioControl bits from PS Remote Play silence the microphone of the controller.
			effects.AudioControl &= dualsensefull.AudioControlOutputPath
			report := effects.OutputReport()
			payload := report[dualsense.OutputReportOffsetUpdateFlags1:]
			if !gamepad.SendEffect(unsafe.Pointer(&payload[0]), len(payload)) {
				slog.Error("Sending the effects to the controller failed", "error", sdl.GetError())
			}
		}
	}
}

func forwardAudio(ctx context.Context, gamepad *sdl.Gamepad, host string) error {
	containerID := gamepad.ContainerID()
	var playbackID, recordingID sdl.AudioDeviceID
	for _, id := range sdl.GetAudioPlaybackDevices() {
		if sdl.GetAudioDeviceContainerID(id) != containerID {
			continue
		}
		if strings.Contains(sdl.GetAudioDeviceName(id), "DualSense") {
			playbackID = id
		}
	}
	for _, id := range sdl.GetAudioRecordingDevices() {
		if sdl.GetAudioDeviceContainerID(id) != containerID {
			continue
		}
		if strings.Contains(sdl.GetAudioDeviceName(id), "DualSense") {
			recordingID = id
		}
	}
	if playbackID == 0 || recordingID == 0 {
		return errors.New("found no playback and recording device of the controller")
	}

	playback, err := sdl.OpenAudioDeviceStream(playbackID, &sdl.AudioSpec{
		Format:   sdl.AudioFormatS16LE,
		Channels: dualsensefull.AudioOutputChannels,
		Freq:     dualsensefull.AudioSampleRate,
	})
	if err != nil {
		return err
	}
	recording, err := sdl.OpenAudioDeviceStream(recordingID, &sdl.AudioSpec{
		Format:   sdl.AudioFormatS16LE,
		Channels: dualsensefull.AudioMicrophoneChannels,
		Freq:     dualsensefull.AudioSampleRate,
	})
	if err != nil {
		playback.Destroy()
		return err
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(audioPort))))
	if err == nil {
		err = playback.ResumeDevice()
	}
	if err == nil {
		err = recording.ResumeDevice()
	}
	if err != nil {
		playback.Destroy()
		recording.Destroy()
		if conn != nil {
			_ = conn.Close()
		}
		return err
	}
	context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})

	go func() {
		defer playback.Destroy()
		buf := make([]byte, dualsensefull.AudioOutputPacketSize)
		for {
			_, err := io.ReadFull(conn, buf)
			if err != nil {
				return
			}
			_ = playback.PutData(buf)
		}
	}()
	go func() {
		defer recording.Destroy()
		buf := make([]byte, dualsensefull.AudioMicrophonePacketSize)
		for ctx.Err() == nil {
			n := recording.GetData(buf)
			if n <= 0 {
				time.Sleep(time.Millisecond)
				continue
			}
			_, err := conn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	slog.Info(
		"Started the audio passthrough",
		"playback", sdl.GetAudioDeviceName(playbackID),
		"recording", sdl.GetAudioDeviceName(recordingID),
	)
	return nil
}
