using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Xbone;


/// <summary>
/// Size in bytes of Xbone output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class Xbone
{
    public const int OutputSize = 6;
}



/// <summary>
/// Button constants for Xbone device.
/// </summary>
[Flags]
public enum Button : uint
{
    Menu = 0x4,
    View = 0x8,
    A = 0x10,
    B = 0x20,
    X = 0x40,
    Y = 0x80,
    DPadUp = 0x100,
    DPadDown = 0x200,
    DPadLeft = 0x400,
    DPadRight = 0x800,
    LShoulder = 0x1000,
    RShoulder = 0x2000,
    LThumb = 0x4000,
    RThumb = 0x8000,
    Guide = 0x10000,
}


/// <summary>
/// GIP_ constants for Xbone device.
/// </summary>
[Flags]
public enum GIP_ : byte
{
    CMD_ACKNOWLEDGE = 0x1,
    CMD_ANNOUNCE = 0x2,
    CMD_STATUS = 0x3,
    CMD_IDENTIFY = 0x4,
    CMD_POWER = 0x5,
    CMD_AUTHENTICATE = 0x6,
    CMD_VIRTUAL_KEY = 0x7,
    CMD_RUMBLE = 0x9,
    CMD_LED = 0xA,
    CMD_INPUT = 0x20,
    OPT_ACKNOWLEDGE = 0x10,
    OPT_INTERNAL = 0x20,
}


/// <summary>
/// ID constants for Xbone device.
/// </summary>
[Flags]
public enum ID : ushort
{
    Vendor = 0x45E,
    Product = 0x2D1,
    ProductS = 0x2EA,
}



