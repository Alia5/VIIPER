using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Dualshock4;


/// <summary>
/// Size in bytes of Dualshock4 output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class Dualshock4
{

    public const int OutputSize = 8;

    public const double AccelCountsPerMS2 = 835.070000;
    public const string DefaultBatteryStatus = "BatteryChargingFlag | BatteryFullyCharged";
    public const string DefaultBoardString = "JDM-055";
    public const string DefaultSerialString = "1111020BF619A500";
    public const double DefaultTemperature = 28.000000;
    public const double DefaultVoltage = 4.200000;
    public const double GyroCountsPerDps = 16.000000;
    public const double StandardGravityMS2 = 9.810000;

}



/// <summary>
/// Battery constants for Dualshock4 device.
/// </summary>
[Flags]
public enum Battery : byte
{
    LevelMask = 0xF,
    ChargingFlag = 0x10,
    FullyCharged = 0xA,
}


/// <summary>
/// Button constants for Dualshock4 device.
/// </summary>
[Flags]
public enum Button : ushort
{
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


/// <summary>
/// D constants for Dualshock4 device.
/// </summary>
[Flags]
public enum D : byte
{
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


/// <summary>
/// Default constants for Dualshock4 device.
/// </summary>
[Flags]
public enum Default : short
{
    VID = 0x54C,
    PID = 0x9CC,
    AccelXRaw = 0x0,
    AccelYRaw = 0x0,
    AccelZRaw = -8192,
    LedRed = 0x0,
    LedGreen = 0x0,
    LedBlue = 0x40,
}


/// <summary>
/// Touchpad constants for Dualshock4 device.
/// </summary>
[Flags]
public enum Touchpad : ushort
{
    MinX = 0x0,
    MaxX = 0x780,
    MinY = 0x0,
    MaxY = 0x3AE,
}


/// <summary>
/// Update constants for Dualshock4 device.
/// </summary>
[Flags]
public enum Update : byte
{
    FlagRumble = 0x1,
    FlagLED = 0x2,
    FlagLEDBlink = 0x4,
}



