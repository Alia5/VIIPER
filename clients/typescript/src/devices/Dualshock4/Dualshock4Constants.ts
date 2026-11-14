// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Dualshock4 constants and maps


export enum Battery {
  LevelMask = 0xF,
  ChargingFlag = 0x10,
  FullyCharged = 0xA,
}

export enum Button {
  Square = 0x10,
  Cross = 0x20,
  Circle = 0x40,
  Triangle = 0x80,
  L1 = 0x100,
  R1 = 0x200,
  L2 = 0x400,
  R2 = 0x800,
  Share = 0x1000,
  Options = 0x2000,
  L3 = 0x4000,
  R3 = 0x8000,
  PS = 0x1,
  TouchpadClick = 0x2,
  PSUSB = 0x1,
  TouchpadClickUSB = 0x2,
}

export enum D {
  PadMask = 0xF,
  PadUSBUp = 0x0,
  PadUSBUpRight = 0x1,
  PadUSBRight = 0x2,
  PadUSBDownRight = 0x3,
  PadUSBDown = 0x4,
  PadUSBDownLeft = 0x5,
  PadUSBLeft = 0x6,
  PadUSBUpLeft = 0x7,
  PadUSBNeutral = 0x8,
  PadUp = 0x1,
  PadDown = 0x2,
  PadLeft = 0x4,
  PadRight = 0x8,
}

export enum Default {
  VID = 0x54C,
  PID = 0x9CC,
  AccelXRaw = 0x0,
  AccelYRaw = 0x0,
  AccelZRaw = -8192,
  LedRed = 0x0,
  LedGreen = 0x0,
  LedBlue = 0x40,
}

export enum Touchpad {
  MinX = 0x0,
  MaxX = 0x780,
  MinY = 0x0,
  MaxY = 0x3AE,
}

export enum Update {
  FlagRumble = 0x1,
  FlagLED = 0x2,
  FlagLEDBlink = 0x4,
}



export const AccelCountsPerMS2: number = 835.070000;

export const DefaultBatteryStatus: string = 'BatteryChargingFlag | BatteryFullyCharged';

export const DefaultBoardString: string = 'JDM-055';

export const DefaultSerialString: string = '1111020BF619A500';

export const DefaultTemperature: number = 28.000000;

export const DefaultVoltage: number = 4.200000;

export const GyroCountsPerDps: number = 16.000000;

export const StandardGravityMS2: number = 9.810000;



