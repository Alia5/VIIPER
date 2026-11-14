use super::constants::*;
use super::trigger_effect::TriggerEffect;
use crate::wire::DeviceOutput;

const TRIGGERS_OFFSET: usize = 9;

#[derive(Debug, Clone, Default)]
pub struct DualsenseOutput {
    pub update_flags1: u8,
    pub update_flags2: u8,
    pub update_flags3: u8,
    pub rumble_small: u8,
    pub rumble_large: u8,
    pub led_red: u8,
    pub led_green: u8,
    pub led_blue: u8,
    pub player_leds: u8,
    pub trigger_r2_mode: u8,
    pub trigger_r2_params: [u8; TRIGGER_PARAMS_SIZE],
    pub trigger_l2_mode: u8,
    pub trigger_l2_params: [u8; TRIGGER_PARAMS_SIZE],
}

impl DualsenseOutput {
    pub fn trigger_r2(&self) -> TriggerEffect {
        TriggerEffect::decode(self.trigger_r2_mode, self.trigger_r2_params)
    }

    pub fn trigger_l2(&self) -> TriggerEffect {
        TriggerEffect::decode(self.trigger_l2_mode, self.trigger_l2_params)
    }
}

impl DeviceOutput for DualsenseOutput {
    fn from_bytes(buf: &[u8]) -> Result<Self, crate::error::ViiperError> {
        if buf.len() < OUTPUT_SIZE {
            return Err(crate::error::ViiperError::UnexpectedResponse(
                "buffer too short".into(),
            ));
        }
        let triggers = &buf[TRIGGERS_OFFSET..];
        let mut trigger_r2_params = [0u8; TRIGGER_PARAMS_SIZE];
        trigger_r2_params.copy_from_slice(&triggers[1..1 + TRIGGER_PARAMS_SIZE]);
        let mut trigger_l2_params = [0u8; TRIGGER_PARAMS_SIZE];
        trigger_l2_params.copy_from_slice(&triggers[2 + TRIGGER_PARAMS_SIZE..2 + 2 * TRIGGER_PARAMS_SIZE]);

        Ok(Self {
            update_flags1: buf[0],
            update_flags2: buf[1],
            update_flags3: buf[2],
            rumble_small: buf[3],
            rumble_large: buf[4],
            led_red: buf[5],
            led_green: buf[6],
            led_blue: buf[7],
            player_leds: buf[8],
            trigger_r2_mode: triggers[0],
            trigger_r2_params,
            trigger_l2_mode: triggers[1 + TRIGGER_PARAMS_SIZE],
            trigger_l2_params,
        })
    }
}
