using System.Diagnostics.CodeAnalysis;
using System.IO;
using Viiper.Client.Devices.Dualsense;

namespace Viiper.Client.Devices.DualsenseFull;

public class DualsenseFullOutput : DualsenseOutput, IBinarySerializable
{
    public required byte TriggerR2Mode { get; set; }
    public required byte[] TriggerR2Params { get; set; }
    public required byte TriggerL2Mode { get; set; }
    public required byte[] TriggerL2Params { get; set; }

    public DualsenseFullOutput()
    {
    }

    [SetsRequiredMembers]
    private DualsenseFullOutput(DualsenseOutput output, byte triggerR2Mode, byte[] triggerR2Params, byte triggerL2Mode, byte[] triggerL2Params)
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
        TriggerR2Mode = triggerR2Mode;
        TriggerR2Params = triggerR2Params;
        TriggerL2Mode = triggerL2Mode;
        TriggerL2Params = triggerL2Params;
    }

    public TriggerEffect TriggerR2() => TriggerEffect.Decode(TriggerR2Mode, TriggerR2Params);

    public TriggerEffect TriggerL2() => TriggerEffect.Decode(TriggerL2Mode, TriggerL2Params);

    public new void Write(BinaryWriter writer)
    {
        base.Write(writer);
        writer.Write(TriggerR2Mode);
        writer.Write(TriggerR2Params, 0, DualsenseFull.TriggerParamsSize);
        writer.Write(TriggerL2Mode);
        writer.Write(TriggerL2Params, 0, DualsenseFull.TriggerParamsSize);
    }

    public new static DualsenseFullOutput Read(BinaryReader reader)
    {
        var output = DualsenseOutput.Read(reader);
        var triggerR2Mode = reader.ReadByte();
        var triggerR2Params = reader.ReadBytes(DualsenseFull.TriggerParamsSize);
        var triggerL2Mode = reader.ReadByte();
        var triggerL2Params = reader.ReadBytes(DualsenseFull.TriggerParamsSize);

        return new DualsenseFullOutput(output, triggerR2Mode, triggerR2Params, triggerL2Mode, triggerL2Params);
    }
}
