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

    pub fn output_report(&self) -> [u8; OUTPUT_REPORT_SIZE as usize] {
        let mut report = [0u8; OUTPUT_REPORT_SIZE as usize];
        report[0] = REPORT_ID_OUTPUT;
        report[OUTPUT_REPORT_OFFSET_UPDATE_FLAGS1] = self.update_flags1;
        report[OUTPUT_REPORT_OFFSET_UPDATE_FLAGS2] = self.update_flags2;
        report[OUTPUT_REPORT_OFFSET_UPDATE_FLAGS3] = self.update_flags3;
        report[OUTPUT_REPORT_OFFSET_RUMBLE_SMALL] = self.rumble_small;
        report[OUTPUT_REPORT_OFFSET_RUMBLE_LARGE] = self.rumble_large;
        report[OUTPUT_REPORT_OFFSET_LED_RED] = self.led_red;
        report[OUTPUT_REPORT_OFFSET_LED_GREEN] = self.led_green;
        report[OUTPUT_REPORT_OFFSET_LED_BLUE] = self.led_blue;
        report[OUTPUT_REPORT_OFFSET_PLAYER_LEDS] = self.player_leds;
        report[OUTPUT_REPORT_OFFSET_TRIGGER_R2_MODE] = self.trigger_r2_mode;
        report[OUTPUT_REPORT_OFFSET_TRIGGER_R2_PARAMS..OUTPUT_REPORT_OFFSET_TRIGGER_R2_PARAMS + TRIGGER_PARAMS_SIZE]
            .copy_from_slice(&self.trigger_r2_params);
        report[OUTPUT_REPORT_OFFSET_TRIGGER_L2_MODE] = self.trigger_l2_mode;
        report[OUTPUT_REPORT_OFFSET_TRIGGER_L2_PARAMS..OUTPUT_REPORT_OFFSET_TRIGGER_L2_PARAMS + TRIGGER_PARAMS_SIZE]
            .copy_from_slice(&self.trigger_l2_params);
        report
    }
}

pub fn gyro_dps_to_raw(dps: f64) -> i16 {
    (dps * GYRO_COUNTS_PER_DPS).round().clamp(i16::MIN as f64, i16::MAX as f64) as i16
}

pub fn accel_ms2_to_raw(ms2: f64) -> i16 {
    (ms2 * ACCEL_COUNTS_PER_MS2).round().clamp(i16::MIN as f64, i16::MAX as f64) as i16
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
