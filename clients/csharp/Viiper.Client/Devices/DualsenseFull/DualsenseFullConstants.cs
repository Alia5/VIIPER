namespace Viiper.Client.Devices.DualsenseFull;

public static class DualsenseFull
{
    public const string DeviceType = "dualsense_full";
    public const string DeviceTypeEdge = "dualsenseedge_full";

    public const int OutputSize = 31;
    public const int TriggerParamsSize = 10;
    public const int TriggerZoneCount = 10;
}

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
