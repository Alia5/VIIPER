export const DeviceType = 'dualsense_full';
export const DeviceTypeEdge = 'dualsenseedge_full';

export const TriggerParamsSize = 10;
export const TriggerZoneCount = 10;
export const OutputStateSize = 31;

export enum TriggerMode {
  SimpleFeedback = 0x01,
  SimpleWeapon = 0x02,
  Off = 0x05,
  SimpleVibration = 0x06,
  LimitedFeedback = 0x11,
  LimitedWeapon = 0x12,
  Feedback = 0x21,
  Bow = 0x22,
  Galloping = 0x23,
  Weapon = 0x25,
  Vibration = 0x26,
  Machine = 0x27,
  Calibration = 0xFC,
  DebugFD = 0xFD,
  DebugFE = 0xFE,
}
