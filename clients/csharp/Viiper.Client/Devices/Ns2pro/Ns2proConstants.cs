using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Ns2pro;


/// <summary>
/// Size in bytes of Ns2pro output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class Ns2pro
{

    public const int OutputSize = 34;

    public const string DefaultSerial = "VIIPER-NS2PRO-00";
    public const string DefaultSerialEnding = "00";

}



/// <summary>
/// Button constants for Ns2pro device.
/// </summary>
[Flags]
public enum Button : uint
{
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


/// <summary>
/// Default constants for Ns2pro device.
/// </summary>
[Flags]
public enum Default : ushort
{
    VID = 0x57E,
    PID = 0x2069,
    BatteryVolts = 0xED8,
}


/// <summary>
/// Endpoint constants for Ns2pro device.
/// </summary>
[Flags]
public enum Endpoint : byte
{
    HIDIn = 0x81,
    HIDOut = 0x1,
    BulkOut = 0x2,
    BulkIn = 0x82,
}


/// <summary>
/// Feature constants for Ns2pro device.
/// </summary>
[Flags]
public enum Feature : byte
{
    Buttons = 0x1,
    Sticks = 0x2,
    IMU = 0x4,
    Mouse = 0x10,
    Rumble = 0x20,
}


/// <summary>
/// Output constants for Ns2pro device.
/// </summary>
[Flags]
public enum Output : byte
{
    FlagRumble = 0x1,
    FlagLED = 0x2,
    ReportSize = 0x40,
    RumbleSize = 0x20,
    WireSize = 0x22,
}


/// <summary>
/// Report constants for Ns2pro device.
/// </summary>
[Flags]
public enum Report : byte
{
    IDCommon = 0x5,
    IDPro = 0x9,
    IDOutput = 0x2,
}


/// <summary>
/// Stick constants for Ns2pro device.
/// </summary>
[Flags]
public enum Stick : ushort
{
    Min = 0x0,
    Center = 0x800,
    Max = 0xFFF,
}



