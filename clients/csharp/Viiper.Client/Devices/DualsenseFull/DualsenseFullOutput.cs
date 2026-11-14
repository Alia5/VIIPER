using System.Diagnostics.CodeAnalysis;
using System.IO;
using Viiper.Client.Devices.Dualsense;

namespace Viiper.Client.Devices.DualsenseFull;

public class DualsenseFullOutput : DualsenseOutput, IBinarySerializable
{
    public required byte Speakervolume { get; set; }
    public required byte Microphonevolume { get; set; }
    public required byte Audiocontrol { get; set; }
    public required byte Micmuteled { get; set; }
    public required byte Powersavecontrol { get; set; }
    public required byte Audiocontrol2 { get; set; }

    public DualsenseFullOutput()
    {
    }

    [SetsRequiredMembers]
    private DualsenseFullOutput(DualsenseOutput output, BinaryReader reader)
    {
        Updateflags1 = output.Updateflags1;
        Updateflags2 = output.Updateflags2;
        Updateflags3 = output.Updateflags3;
        Rumblesmall = output.Rumblesmall;
        Rumblelarge = output.Rumblelarge;
        Ledred = output.Ledred;
        Ledgreen = output.Ledgreen;
        Ledblue = output.Ledblue;
        Playerleds = output.Playerleds;
        TriggerR2Mode = output.TriggerR2Mode;
        TriggerR2Params = output.TriggerR2Params;
        TriggerL2Mode = output.TriggerL2Mode;
        TriggerL2Params = output.TriggerL2Params;
        Speakervolume = reader.ReadByte();
        Microphonevolume = reader.ReadByte();
        Audiocontrol = reader.ReadByte();
        Micmuteled = reader.ReadByte();
        Powersavecontrol = reader.ReadByte();
        Audiocontrol2 = reader.ReadByte();
    }

    public new byte[] OutputReport()
    {
        var report = base.OutputReport();
        report[DualsenseFull.OutputReportOffsetSpeakerVolume] = Speakervolume;
        report[DualsenseFull.OutputReportOffsetMicrophoneVolume] = Microphonevolume;
        report[DualsenseFull.OutputReportOffsetAudioControl] = Audiocontrol;
        report[DualsenseFull.OutputReportOffsetMicMuteLed] = Micmuteled;
        report[DualsenseFull.OutputReportOffsetPowerSaveControl] = Powersavecontrol;
        report[DualsenseFull.OutputReportOffsetAudioControl2] = Audiocontrol2;
        return report;
    }

    public new void Write(BinaryWriter writer)
    {
        base.Write(writer);
        writer.Write(Speakervolume);
        writer.Write(Microphonevolume);
        writer.Write(Audiocontrol);
        writer.Write(Micmuteled);
        writer.Write(Powersavecontrol);
        writer.Write(Audiocontrol2);
    }

    public new static DualsenseFullOutput Read(BinaryReader reader)
    {
        return new DualsenseFullOutput(DualsenseOutput.Read(reader), reader);
    }
}
