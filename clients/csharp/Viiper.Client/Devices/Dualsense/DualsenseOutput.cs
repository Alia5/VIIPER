using System;
using System.IO;

namespace Viiper.Client.Devices.Dualsense;

/// <summary>
/// Wire protocol Output message for Dualsense device.
/// </summary>
public class DualsenseOutput : IBinarySerializable
{
    public required byte Updateflags1 { get; set; }
    public required byte Updateflags2 { get; set; }
    public required byte Updateflags3 { get; set; }
    public required byte Rumblesmall { get; set; }
    public required byte Rumblelarge { get; set; }
    public required byte Ledred { get; set; }
    public required byte Ledgreen { get; set; }
    public required byte Ledblue { get; set; }
    public required byte Playerleds { get; set; }
    public required byte TriggerR2Mode { get; set; }
    public required byte[] TriggerR2Params { get; set; }
    public required byte TriggerL2Mode { get; set; }
    public required byte[] TriggerL2Params { get; set; }

    public TriggerEffect TriggerR2() => TriggerEffect.Decode(TriggerR2Mode, TriggerR2Params);

    public TriggerEffect TriggerL2() => TriggerEffect.Decode(TriggerL2Mode, TriggerL2Params);

    public byte[] OutputReport()
    {
        var report = new byte[Dualsense.OutputReportSize];
        report[0] = Dualsense.ReportIDOutput;
        report[Dualsense.OutputReportOffsetUpdateFlags1] = Updateflags1;
        report[Dualsense.OutputReportOffsetUpdateFlags2] = Updateflags2;
        report[Dualsense.OutputReportOffsetUpdateFlags3] = Updateflags3;
        report[Dualsense.OutputReportOffsetRumbleSmall] = Rumblesmall;
        report[Dualsense.OutputReportOffsetRumbleLarge] = Rumblelarge;
        report[Dualsense.OutputReportOffsetLedRed] = Ledred;
        report[Dualsense.OutputReportOffsetLedGreen] = Ledgreen;
        report[Dualsense.OutputReportOffsetLedBlue] = Ledblue;
        report[Dualsense.OutputReportOffsetPlayerLeds] = Playerleds;
        report[Dualsense.OutputReportOffsetTriggerR2Mode] = TriggerR2Mode;
        Array.Copy(TriggerR2Params, 0, report, Dualsense.OutputReportOffsetTriggerR2Params, Dualsense.TriggerParamsSize);
        report[Dualsense.OutputReportOffsetTriggerL2Mode] = TriggerL2Mode;
        Array.Copy(TriggerL2Params, 0, report, Dualsense.OutputReportOffsetTriggerL2Params, Dualsense.TriggerParamsSize);
        return report;
    }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Updateflags1);
        writer.Write(Updateflags2);
        writer.Write(Updateflags3);
        writer.Write(Rumblesmall);
        writer.Write(Rumblelarge);
        writer.Write(Ledred);
        writer.Write(Ledgreen);
        writer.Write(Ledblue);
        writer.Write(Playerleds);
        writer.Write(TriggerR2Mode);
        writer.Write(TriggerR2Params, 0, Dualsense.TriggerParamsSize);
        writer.Write(TriggerL2Mode);
        writer.Write(TriggerL2Params, 0, Dualsense.TriggerParamsSize);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static DualsenseOutput Read(BinaryReader reader)
    {
        return new DualsenseOutput
        {
            Updateflags1 = reader.ReadByte(),
            Updateflags2 = reader.ReadByte(),
            Updateflags3 = reader.ReadByte(),
            Rumblesmall = reader.ReadByte(),
            Rumblelarge = reader.ReadByte(),
            Ledred = reader.ReadByte(),
            Ledgreen = reader.ReadByte(),
            Ledblue = reader.ReadByte(),
            Playerleds = reader.ReadByte(),
            TriggerR2Mode = reader.ReadByte(),
            TriggerR2Params = reader.ReadBytes(Dualsense.TriggerParamsSize),
            TriggerL2Mode = reader.ReadByte(),
            TriggerL2Params = reader.ReadBytes(Dualsense.TriggerParamsSize),
        };
    }
}
