use super::constants::*;

const BITS_PER_STRENGTH: usize = 3;
const STRENGTH_MASK: u8 = 0x07;
const FREQUENCY_PARAM: usize = 8;

#[derive(Debug, Clone, Default)]
pub struct TriggerEffect {
    pub mode: u8,
    pub params: [u8; TRIGGER_PARAMS_SIZE],
    pub zone_strengths: [u8; TRIGGER_ZONE_COUNT],
    pub start_zone: u8,
    pub end_zone: u8,
    pub start_position: u8,
    pub end_position: u8,
    pub strength: u8,
    pub snap_force: u8,
    pub amplitude_a: u8,
    pub amplitude_b: u8,
    pub first_foot: u8,
    pub second_foot: u8,
    pub frequency: u8,
    pub period: u8,
}

impl TriggerEffect {
    pub fn decode(mode: u8, params: [u8; TRIGGER_PARAMS_SIZE]) -> Self {
        let mut effect = TriggerEffect {
            mode,
            params,
            ..Default::default()
        };
        let zones = u16::from_le_bytes([params[0], params[1]]);

        match mode {
            TRIGGER_MODE_FEEDBACK | TRIGGER_MODE_VIBRATION => {
                let packed_strengths = u32::from_le_bytes([params[2], params[3], params[4], params[5]]);
                for zone in 0..TRIGGER_ZONE_COUNT {
                    if zones & (1 << zone) != 0 {
                        effect.zone_strengths[zone] = ((packed_strengths >> (BITS_PER_STRENGTH * zone)) as u8 & STRENGTH_MASK) + 1;
                    }
                }
                if mode == TRIGGER_MODE_VIBRATION {
                    effect.frequency = params[FREQUENCY_PARAM];
                }
            }
            TRIGGER_MODE_WEAPON | TRIGGER_MODE_BOW | TRIGGER_MODE_GALLOPING | TRIGGER_MODE_MACHINE => {
                if zones != 0 {
                    effect.start_zone = zones.trailing_zeros() as u8;
                    effect.end_zone = (u16::BITS - 1 - zones.leading_zeros()) as u8;
                }
                let first = params[2] & STRENGTH_MASK;
                let second = (params[2] >> BITS_PER_STRENGTH) & STRENGTH_MASK;
                match mode {
                    TRIGGER_MODE_WEAPON => {
                        effect.strength = first + 1;
                    }
                    TRIGGER_MODE_BOW => {
                        effect.strength = first + 1;
                        effect.snap_force = second + 1;
                    }
                    TRIGGER_MODE_GALLOPING => {
                        effect.second_foot = first;
                        effect.first_foot = second;
                        effect.frequency = params[3];
                    }
                    _ => {
                        effect.amplitude_a = first;
                        effect.amplitude_b = second;
                        effect.frequency = params[3];
                        effect.period = params[4];
                    }
                }
            }
            TRIGGER_MODE_SIMPLE_FEEDBACK | TRIGGER_MODE_LIMITED_FEEDBACK => {
                effect.start_position = params[0];
                effect.strength = params[1];
            }
            TRIGGER_MODE_SIMPLE_WEAPON | TRIGGER_MODE_LIMITED_WEAPON => {
                effect.start_position = params[0];
                effect.end_position = params[1];
                effect.strength = params[2];
            }
            TRIGGER_MODE_SIMPLE_VIBRATION => {
                effect.frequency = params[0];
                effect.strength = params[1];
                effect.start_position = params[2];
            }
            _ => {}
        }
        effect
    }
}
