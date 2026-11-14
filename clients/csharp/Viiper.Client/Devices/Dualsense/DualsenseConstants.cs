using System;
using System.Collections.Generic;

namespace Viiper.Client.Devices.Dualsense;


/// <summary>
/// Size in bytes of Dualsense output (server-to-client) messages.
/// Use this constant to allocate buffers for reading device output.
/// </summary>
public static class Dualsense
{

    public const int OutputSize = 31;
    public const int TriggerParamsSize = 10;
    public const int TriggerZoneCount = 10;

    public const double AccelCountsPerMS2 = 835.070000;
    public const string DefaultBatteryStatus = "BatteryFullyCharged";
    public const string DefaultBoardStringDS = "BDM-050";
    public const string DefaultBoardStringEdge = "HMB-010";
    public const string DefaultMACAddressDS = "A5:FA:9C:CF:92:00";
    public const string DefaultMACAddressDSEdge = "A5:FE:9C:CF:92:00";
    public const string DefaultSerialNumberDS = "E55700GTD1190A500";
    public const string DefaultSerialNumberDSEdge = "E55E00GTD1190A500";
    public const string DefaultShellColor = "ShellColorBlack";
    public const double DefaultTemperature = 28.000000;
    public const double DefaultVoltage = 3.800000;
    public const double GyroCountsPerDps = 16.384000;
    public const string ShellColorAnniversary30th = "30";
    public const string ShellColorAstroBot = "Z3";
    public const string ShellColorAstroBotJoyfulEdition = "ZC";
    public const string ShellColorBlack = "01";
    public const string ShellColorChromaIndigo = "11";
    public const string ShellColorChromaPearl = "12";
    public const string ShellColorChromaTeal = "10";
    public const string ShellColorCobaltBlue = "09";
    public const string ShellColorCosmicRed = "02";
    public const string ShellColorFortnite = "Z4";
    public const string ShellColorGalacticPurple = "04";
    public const string ShellColorGenshinImpact = "ZE";
    public const string ShellColorGhostOfYotei = "Z7";
    public const string ShellColorGodOfWarRagnarok = "Z1";
    public const string ShellColorGreyCamouflage = "06";
    public const string ShellColorIconBlueLimitedEdition = "ZB";
    public const string ShellColorMonsterHunterWilds = "Z5";
    public const string ShellColorNovaPink = "03";
    public const string ShellColorSpiderMan2 = "Z2";
    public const string ShellColorStarlightBlue = "05";
    public const string ShellColorSterlingSilver = "08";
    public const string ShellColorTheLastOfUs = "Z6";
    public const string ShellColorVolcanicRed = "07";
    public const string ShellColorWhite = "00";

}



/// <summary>
/// Button constants for Dualsense device.
/// </summary>
[Flags]
public enum Button : uint
{
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


/// <summary>
/// D constants for Dualsense device.
/// </summary>
[Flags]
public enum D : byte
{
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


/// <summary>
/// Default constants for Dualsense device.
/// </summary>
[Flags]
public enum Default : short
{
    VID = 0x54C,
    PIDDSEdge = 0xDF2,
    PIDDS = 0xCE6,
    AccelXRaw = 0x0,
    AccelYRaw = 0x0,
    AccelZRaw = -8192,
}


/// <summary>
/// Touchpad constants for Dualsense device.
/// </summary>
[Flags]
public enum Touchpad : ushort
{
    MinX = 0x0,
    MinY = 0x0,
    MaxX = 0x780,
    MaxY = 0x438,
}


/// <summary>
/// Update constants for Dualsense device.
/// </summary>
[Flags]
public enum Update : byte
{
    Flag1Rumble = 0x1,
    Flag1HapticsSelect = 0x2,
    Flag1TriggerR2 = 0x4,
    Flag1TriggerL2 = 0x8,
    Flag2Led = 0x4,
    Flag2PlayerLeds = 0x10,
    Flag3ImprovedRumble = 0x4,
}


/// <summary>
/// Adaptive trigger mode constants for Dualsense device.
/// </summary>
public enum TriggerMode : byte
{
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



