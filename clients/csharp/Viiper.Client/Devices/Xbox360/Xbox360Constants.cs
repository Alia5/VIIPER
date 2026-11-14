using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Xbox360;


/// <summary>
/// Size in bytes of Xbox360 output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class Xbox360
{

    public const int OutputSize = 2;


}



/// <summary>
/// Button constants for Xbox360 device.
/// </summary>
[Flags]
public enum Button : ushort
{
    DPadUp = 0x1,
    DPadDown = 0x2,
    DPadLeft = 0x4,
    DPadRight = 0x8,
    Start = 0x10,
    Back = 0x20,
    LThumb = 0x40,
    RThumb = 0x80,
    LShoulder = 0x100,
    RShoulder = 0x200,
    Guide = 0x400,
    A = 0x1000,
    B = 0x2000,
    X = 0x4000,
    Y = 0x8000,
}



