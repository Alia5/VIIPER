using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Mouse;



/// <summary>
/// Btn constants for Mouse device.
/// </summary>
[Flags]
public enum Btn : byte
{
    Left = 0x1,
    Right = 0x2,
    Middle = 0x4,
    Back = 0x8,
    Forward = 0x10,
}



