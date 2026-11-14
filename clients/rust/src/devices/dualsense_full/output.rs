use super::constants::*;
use super::trigger_effect::TriggerEffect;
use crate::devices::dualsense::DualsenseOutput;
use crate::wire::DeviceOutput;

const BASE_OUTPUT_SIZE: usize = crate::devices::dualsense::OUTPUT_SIZE;

#[derive(Debug, Clone, Default)]
pub struct DualsenseFullOutput {
    pub base: DualsenseOutput,
    pub trigger_r2_mode: u8,
    pub trigger_r2_params: [u8; TRIGGER_PARAMS_SIZE],
    pub trigger_l2_mode: u8,
    pub trigger_l2_params: [u8; TRIGGER_PARAMS_SIZE],
}

impl DualsenseFullOutput {
    pub fn trigger_r2(&self) -> TriggerEffect {
        TriggerEffect::decode(self.trigger_r2_mode, self.trigger_r2_params)
    }

    pub fn trigger_l2(&self) -> TriggerEffect {
        TriggerEffect::decode(self.trigger_l2_mode, self.trigger_l2_params)
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
        let triggers = &buf[BASE_OUTPUT_SIZE..OUTPUT_SIZE];
        let mut trigger_r2_params = [0u8; TRIGGER_PARAMS_SIZE];
        trigger_r2_params.copy_from_slice(&triggers[1..1 + TRIGGER_PARAMS_SIZE]);
        let mut trigger_l2_params = [0u8; TRIGGER_PARAMS_SIZE];
        trigger_l2_params.copy_from_slice(&triggers[2 + TRIGGER_PARAMS_SIZE..]);

        Ok(Self {
            base,
            trigger_r2_mode: triggers[0],
            trigger_r2_params,
            trigger_l2_mode: triggers[1 + TRIGGER_PARAMS_SIZE],
            trigger_l2_params,
        })
    }
}
