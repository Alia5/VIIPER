package dualsensefull

import (
	"io"

	"github.com/Alia5/VIIPER/device/dualsense"
)

// nolint
// viiper:wire dualsense_full s2c updateFlags1:u8 updateFlags2:u8 updateFlags3:u8 rumbleSmall:u8 rumbleLarge:u8 ledRed:u8 ledGreen:u8 ledBlue:u8 playerLeds:u8 triggerR2Mode:u8 triggerR2Params:u8*10 triggerL2Mode:u8 triggerL2Params:u8*10 speakerVolume:u8 microphoneVolume:u8 audioControl:u8 micMuteLed:u8 powerSaveControl:u8 audioControl2:u8
type OutputState struct {
	dualsense.OutputState

	SpeakerVolume    uint8
	MicrophoneVolume uint8
	AudioControl     uint8
	MicMuteLed       uint8
	PowerSaveControl uint8
	AudioControl2    uint8
}

func (f *OutputState) OutputReport() [dualsense.OutputReportSize]byte {
	report := f.OutputState.OutputReport()
	report[OutputReportOffsetSpeakerVolume] = f.SpeakerVolume
	report[OutputReportOffsetMicrophoneVolume] = f.MicrophoneVolume
	report[OutputReportOffsetAudioControl] = f.AudioControl
	report[OutputReportOffsetMicMuteLed] = f.MicMuteLed
	report[OutputReportOffsetPowerSaveControl] = f.PowerSaveControl
	report[OutputReportOffsetAudioControl2] = f.AudioControl2
	return report
}

func (f *OutputState) MarshalBinary() ([]byte, error) {
	b, err := f.OutputState.MarshalBinary()
	if err != nil {
		return nil, err
	}
	b = append(
		b,
		f.SpeakerVolume,
		f.MicrophoneVolume,
		f.AudioControl,
		f.MicMuteLed,
		f.PowerSaveControl,
		f.AudioControl2,
	)
	return b, nil
}

func (f *OutputState) UnmarshalBinary(data []byte) error {
	if len(data) < OutputStateSize {
		return io.ErrUnexpectedEOF
	}
	err := f.OutputState.UnmarshalBinary(data)
	if err != nil {
		return err
	}
	audio := data[dualsense.OutputStateSize:]
	f.SpeakerVolume = audio[0]
	f.MicrophoneVolume = audio[1]
	f.AudioControl = audio[2]
	f.MicMuteLed = audio[3]
	f.PowerSaveControl = audio[4]
	f.AudioControl2 = audio[5]
	return nil
}
