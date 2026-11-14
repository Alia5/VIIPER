package dualsensefull

import (
	"math"

	"github.com/Alia5/VIIPER/device/dualsense"
)

const (
	AudioSampleRate         = 48000
	AudioBytesPerSample     = 2
	AudioOutputChannels     = 4
	AudioMicrophoneChannels = 2

	audioPacketsPerSecond     = 1000
	AudioOutputPacketSize     = AudioSampleRate / audioPacketsPerSecond * AudioOutputChannels * AudioBytesPerSample
	AudioMicrophonePacketSize = AudioSampleRate / audioPacketsPerSecond * AudioMicrophoneChannels * AudioBytesPerSample
)

const OutputStateSize = dualsense.OutputStateSize + 6

const (
	OutputReportOffsetSpeakerVolume    = 6
	OutputReportOffsetMicrophoneVolume = 7
	OutputReportOffsetAudioControl     = 8
	OutputReportOffsetMicMuteLed       = 9
	OutputReportOffsetPowerSaveControl = 10
	OutputReportOffsetAudioControl2    = 38
)

const AudioControlOutputPath uint8 = 0x30

const PowerSaveControlMicrophoneMute uint8 = 0x10

const ConnectStateMicrophoneMuted uint8 = 0x04

const (
	interfaceAudioControl = 0x00
	interfaceCount        = 4

	endpointAudioOut     = 1
	endpointMicrophoneIn = 2

	microphoneFrameSize  = AudioMicrophoneChannels * AudioBytesPerSample
	microphoneBufferSize = 100 * AudioMicrophonePacketSize
	audioOutQueueSize    = 64
	audioReadBufferSize  = 1024 * microphoneFrameSize
)

const (
	requestSetInterface uint8 = 0x0B
	requestGetInterface uint8 = 0x0A

	audioRequestTypeSet uint8 = 0x21
	audioRequestTypeGet uint8 = 0xA1

	audioControlMute   uint8 = 0x01
	audioControlVolume uint8 = 0x02

	audioEntitySpeakerFeatureUnit    uint8 = 0x02
	audioEntityMicrophoneFeatureUnit uint8 = 0x05
)

type audioRequest uint8

const (
	audioRequestSetCur audioRequest = 0x01
	audioRequestGetCur audioRequest = 0x81
	audioRequestGetMin audioRequest = 0x82
	audioRequestGetMax audioRequest = 0x83
	audioRequestGetRes audioRequest = 0x84
)

const (
	audioVolumeDecibel int16 = 256
	audioVolumeSilence int16 = math.MinInt16

	speakerVolumeMin int16 = -100 * audioVolumeDecibel
	speakerVolumeMax int16 = 0
	speakerVolumeRes int16 = audioVolumeDecibel

	microphoneVolumeMin int16 = 0
	microphoneVolumeMax int16 = 48 * audioVolumeDecibel
	microphoneVolumeRes int16 = 0x007A
)
