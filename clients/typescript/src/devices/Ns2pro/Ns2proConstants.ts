// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Ns2pro constants and maps


export enum Button {
  B = 0x1,
  A = 0x2,
  Y = 0x4,
  X = 0x8,
  R = 0x10,
  ZR = 0x20,
  Plus = 0x40,
  RightStick = 0x80,
  Down = 0x100,
  Right = 0x200,
  Left = 0x400,
  Up = 0x800,
  L = 0x1000,
  ZL = 0x2000,
  Minus = 0x4000,
  LeftStick = 0x8000,
  Home = 0x10000,
  Capture = 0x20000,
  GR = 0x40000,
  GL = 0x80000,
  C = 0x100000,
  Headset = 0x200000,
}

export enum Default {
  VID = 0x57E,
  PID = 0x2069,
  BatteryVolts = 0xED8,
}

export enum Endpoint {
  HIDIn = 0x81,
  HIDOut = 0x1,
  BulkOut = 0x2,
  BulkIn = 0x82,
}

export enum Feature {
  Buttons = 0x1,
  Sticks = 0x2,
  IMU = 0x4,
  Mouse = 0x10,
  Rumble = 0x20,
}

export enum Output {
  FlagRumble = 0x1,
  FlagLED = 0x2,
  ReportSize = 0x40,
  RumbleSize = 0x20,
  WireSize = 0x22,
}

export enum Report {
  IDCommon = 0x5,
  IDPro = 0x9,
  IDOutput = 0x2,
}

export enum Stick {
  Min = 0x0,
  Center = 0x800,
  Max = 0xFFF,
}



export const DefaultSerial: string = 'VIIPER-NS2PRO-00';

export const DefaultSerialEnding: string = '00';



