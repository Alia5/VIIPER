using System;
using System.IO;

namespace Viiper.Client.Devices.Xbox360;

/// <summary>
/// Wire protocol Input message for Xbox360 device.
/// </summary>
public class Xbox360Input : IBinarySerializable
{
    public required uint Buttons { get; set; }
    public required byte Lt { get; set; }
    public required byte Rt { get; set; }
    public required short Lx { get; set; }
    public required short Ly { get; set; }
    public required short Rx { get; set; }
    public required short Ry { get; set; }
    public byte[] Reserved { get; set; } = new byte[6];

    public void Write(BinaryWriter writer)
    {
        writer.Write(Buttons);
        writer.Write(Lt);
        writer.Write(Rt);
        writer.Write(Lx);
        writer.Write(Ly);
        writer.Write(Rx);
        writer.Write(Ry);
        for (int i = 0; i < 6; i++)
		{
			writer.Write((Reserved != null && i < Reserved.Length) ? Reserved[i] : default(byte));
		}
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static Xbox360Input Read(BinaryReader reader)
    {
	        var buttons = reader.ReadUInt32();
	        var lt = reader.ReadByte();
	        var rt = reader.ReadByte();
	        var lx = reader.ReadInt16();
	        var ly = reader.ReadInt16();
	        var rx = reader.ReadInt16();
	        var ry = reader.ReadInt16();
	        var reserved = new byte[6];
		for (int i = 0; i < 6; i++)
		{
		    reserved[i] = reader.ReadByte();
		}
	

		return new Xbox360Input
		{
	            Buttons = buttons,
	            Lt = lt,
	            Rt = rt,
	            Lx = lx,
	            Ly = ly,
	            Rx = rx,
	            Ry = ry,
	            Reserved = reserved,
	        };
	    }
}
