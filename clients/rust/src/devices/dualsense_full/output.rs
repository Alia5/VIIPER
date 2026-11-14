use super::constants::*;
use crate::devices::dualsense::DualsenseOutput;
use crate::wire::DeviceOutput;

const BASE_OUTPUT_SIZE: usize = crate::devices::dualsense::OUTPUT_SIZE;

#[derive(Debug, Clone, Default)]
pub struct DualsenseFullOutput {
    pub base: DualsenseOutput,
    pub speaker_volume: u8,
    pub microphone_volume: u8,
    pub audio_control: u8,
    pub mic_mute_led: u8,
    pub power_save_control: u8,
    pub audio_control2: u8,
}

impl DualsenseFullOutput {
    pub fn output_report(&self) -> [u8; crate::devices::dualsense::OUTPUT_REPORT_SIZE as usize] {
        let mut report = self.base.output_report();
        report[OUTPUT_REPORT_OFFSET_SPEAKER_VOLUME] = self.speaker_volume;
        report[OUTPUT_REPORT_OFFSET_MICROPHONE_VOLUME] = self.microphone_volume;
        report[OUTPUT_REPORT_OFFSET_AUDIO_CONTROL] = self.audio_control;
        report[OUTPUT_REPORT_OFFSET_MIC_MUTE_LED] = self.mic_mute_led;
        report[OUTPUT_REPORT_OFFSET_POWER_SAVE_CONTROL] = self.power_save_control;
        report[OUTPUT_REPORT_OFFSET_AUDIO_CONTROL2] = self.audio_control2;
        report
    }
}

impl DeviceOutput for DualsenseFullOutput {
    fn from_bytes(buf: &[u8]) -> Result<Self, crate::error::ViiperError> {
        if buf.len() < OUTPUT_SIZE {
            return Err(crate::error::ViiperError::UnexpectedResponse(
                "buffer too short".into(),
            ));
        }
        let base = DualsenseOutput::from_bytes(buf)?;
        let audio = &buf[BASE_OUTPUT_SIZE..OUTPUT_SIZE];

        Ok(Self {
            base,
            speaker_volume: audio[0],
            microphone_volume: audio[1],
            audio_control: audio[2],
            mic_mute_led: audio[3],
            power_save_control: audio[4],
            audio_control2: audio[5],
        })
    }
}
