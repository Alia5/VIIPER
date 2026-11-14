using System;
using System.IO;

namespace Viiper.Client.Devices.Xbone;

/// <summary>
/// Wire protocol Output message for Xbone device.
/// </summary>
public class XboneOutput : IBinarySerializable
{
    public required byte Left { get; set; }
    public required byte Right { get; set; }
    public required byte Lefttrigger { get; set; }
    public required byte Righttrigger { get; set; }
    public required byte Duration { get; set; }
    public required byte Delay { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Left);
        writer.Write(Right);
        writer.Write(Lefttrigger);
        writer.Write(Righttrigger);
        writer.Write(Duration);
        writer.Write(Delay);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static XboneOutput Read(BinaryReader reader)
    {
        return new XboneOutput
        {
            Left = reader.ReadByte(),
            Right = reader.ReadByte(),
            Lefttrigger = reader.ReadByte(),
            Righttrigger = reader.ReadByte(),
            Duration = reader.ReadByte(),
            Delay = reader.ReadByte(),
        };
    }
}
