pub const DEVICE_TYPE: &str = "dualsense_full";
pub const DEVICE_TYPE_EDGE: &str = "dualsenseedge_full";

pub const OUTPUT_SIZE: usize = 31;
pub const TRIGGER_PARAMS_SIZE: usize = 10;
pub const TRIGGER_ZONE_COUNT: usize = 10;

pub const TRIGGER_MODE_SIMPLE_FEEDBACK: u8 = 0x01;
pub const TRIGGER_MODE_SIMPLE_WEAPON: u8 = 0x02;
pub const TRIGGER_MODE_OFF: u8 = 0x05;
pub const TRIGGER_MODE_SIMPLE_VIBRATION: u8 = 0x06;
pub const TRIGGER_MODE_LIMITED_FEEDBACK: u8 = 0x11;
pub const TRIGGER_MODE_LIMITED_WEAPON: u8 = 0x12;
pub const TRIGGER_MODE_FEEDBACK: u8 = 0x21;
pub const TRIGGER_MODE_BOW: u8 = 0x22;
pub const TRIGGER_MODE_GALLOPING: u8 = 0x23;
pub const TRIGGER_MODE_WEAPON: u8 = 0x25;
pub const TRIGGER_MODE_VIBRATION: u8 = 0x26;
pub const TRIGGER_MODE_MACHINE: u8 = 0x27;
pub const TRIGGER_MODE_CALIBRATION: u8 = 0xFC;
pub const TRIGGER_MODE_DEBUG_FD: u8 = 0xFD;
pub const TRIGGER_MODE_DEBUG_FE: u8 = 0xFE;
