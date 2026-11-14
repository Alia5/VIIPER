using System;
using System.IO;

namespace Viiper.Client.Devices.Xbone;

/// <summary>
/// Wire protocol Input message for Xbone device.
/// </summary>
public class XboneInput : IBinarySerializable
{
    public required uint Buttons { get; set; }
    public required ushort Lt { get; set; }
    public required ushort Rt { get; set; }
    public required short Lx { get; set; }
    public required short Ly { get; set; }
    public required short Rx { get; set; }
    public required short Ry { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Buttons);
        writer.Write(Lt);
        writer.Write(Rt);
        writer.Write(Lx);
        writer.Write(Ly);
        writer.Write(Rx);
        writer.Write(Ry);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static XboneInput Read(BinaryReader reader)
    {
        return new XboneInput
        {
            Buttons = reader.ReadUInt32(),
            Lt = reader.ReadUInt16(),
            Rt = reader.ReadUInt16(),
            Lx = reader.ReadInt16(),
            Ly = reader.ReadInt16(),
            Rx = reader.ReadInt16(),
            Ry = reader.ReadInt16(),
        };
    }
}
