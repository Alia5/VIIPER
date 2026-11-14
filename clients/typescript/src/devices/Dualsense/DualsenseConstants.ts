// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Dualsense constants and maps


export enum Button {
  Square = 0x10,
  Cross = 0x20,
  Circle = 0x40,
  Triangle = 0x80,
  L1 = 0x100,
  R1 = 0x200,
  L2 = 0x400,
  R2 = 0x800,
  Create = 0x1000,
  Options = 0x2000,
  L3 = 0x4000,
  R3 = 0x8000,
  PS = 0x10000,
  Touchpad = 0x20000,
  MicMute = 0x40000,
  EdgeLFn = 0x100000,
  EdgeRFn = 0x200000,
  EdgeL4 = 0x400000,
  EdgeR4 = 0x800000,
}

export enum D {
  PadUp = 0x1,
  PadDown = 0x2,
  PadLeft = 0x4,
  PadRight = 0x8,
  PadUSBUp = 0x0,
  PadUSBUpRight = 0x1,
  PadUSBRight = 0x2,
  PadUSBDownRight = 0x3,
  PadUSBDown = 0x4,
  PadUSBDownLeft = 0x5,
  PadUSBLeft = 0x6,
  PadUSBUpLeft = 0x7,
  PadUSBNeutral = 0x8,
  PadMask = 0xF,
}

export enum Default {
  VID = 0x54C,
  PIDDSEdge = 0xDF2,
  PIDDS = 0xCE6,
  AccelXRaw = 0x0,
  AccelYRaw = 0x0,
  AccelZRaw = -8192,
}

export enum Touchpad {
  MinX = 0x0,
  MinY = 0x0,
  MaxX = 0x780,
  MaxY = 0x438,
}

export enum Update {
  Flag1Rumble = 0x1,
  Flag1HapticsSelect = 0x2,
  Flag1TriggerR2 = 0x4,
  Flag1TriggerL2 = 0x8,
  Flag2Led = 0x4,
  Flag2PlayerLeds = 0x10,
  Flag3ImprovedRumble = 0x4,
}

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

export const TriggerParamsSize = 10;
export const TriggerZoneCount = 10;
export const OutputStateSize = 31;



export const AccelCountsPerMS2: number = 835.070000;

export const DefaultBatteryStatus: string = 'BatteryFullyCharged';

export const DefaultBoardStringDS: string = 'BDM-050';

export const DefaultBoardStringEdge: string = 'HMB-010';

export const DefaultMACAddressDS: string = 'A5:FA:9C:CF:92:00';

export const DefaultMACAddressDSEdge: string = 'A5:FE:9C:CF:92:00';

export const DefaultSerialNumberDS: string = 'E55700GTD1190A500';

export const DefaultSerialNumberDSEdge: string = 'E55E00GTD1190A500';

export const DefaultShellColor: string = 'ShellColorBlack';

export const DefaultTemperature: number = 28.000000;

export const DefaultVoltage: number = 3.800000;

export const GyroCountsPerDps: number = 16.384000;

export const ShellColorAnniversary30th: string = '30';

export const ShellColorAstroBot: string = 'Z3';

export const ShellColorAstroBotJoyfulEdition: string = 'ZC';

export const ShellColorBlack: string = '01';

export const ShellColorChromaIndigo: string = '11';

export const ShellColorChromaPearl: string = '12';

export const ShellColorChromaTeal: string = '10';

export const ShellColorCobaltBlue: string = '09';

export const ShellColorCosmicRed: string = '02';

export const ShellColorFortnite: string = 'Z4';

export const ShellColorGalacticPurple: string = '04';

export const ShellColorGenshinImpact: string = 'ZE';

export const ShellColorGhostOfYotei: string = 'Z7';

export const ShellColorGodOfWarRagnarok: string = 'Z1';

export const ShellColorGreyCamouflage: string = '06';

export const ShellColorIconBlueLimitedEdition: string = 'ZB';

export const ShellColorMonsterHunterWilds: string = 'Z5';

export const ShellColorNovaPink: string = '03';

export const ShellColorSpiderMan2: string = 'Z2';

export const ShellColorStarlightBlue: string = '05';

export const ShellColorSterlingSilver: string = '08';

export const ShellColorTheLastOfUs: string = 'Z6';

export const ShellColorVolcanicRed: string = '07';

export const ShellColorWhite: string = '00';



