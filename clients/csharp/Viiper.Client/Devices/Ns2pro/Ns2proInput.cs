using System;
using System.IO;

namespace Viiper.Client.Devices.Ns2pro;

/// <summary>
/// Wire protocol Input message for Ns2pro device.
/// </summary>
public class Ns2proInput : IBinarySerializable
{
    public required uint Buttons { get; set; }
    public required ushort Lx { get; set; }
    public required ushort Ly { get; set; }
    public required ushort Rx { get; set; }
    public required ushort Ry { get; set; }
    public required short Accelx { get; set; }
    public required short Accely { get; set; }
    public required short Accelz { get; set; }
    public required short Gyrox { get; set; }
    public required short Gyroy { get; set; }
    public required short Gyroz { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Buttons);
        writer.Write(Lx);
        writer.Write(Ly);
        writer.Write(Rx);
        writer.Write(Ry);
        writer.Write(Accelx);
        writer.Write(Accely);
        writer.Write(Accelz);
        writer.Write(Gyrox);
        writer.Write(Gyroy);
        writer.Write(Gyroz);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static Ns2proInput Read(BinaryReader reader)
    {
	        var buttons = reader.ReadUInt32();
	        var lx = reader.ReadUInt16();
	        var ly = reader.ReadUInt16();
	        var rx = reader.ReadUInt16();
	        var ry = reader.ReadUInt16();
	        var accelx = reader.ReadInt16();
	        var accely = reader.ReadInt16();
	        var accelz = reader.ReadInt16();
	        var gyrox = reader.ReadInt16();
	        var gyroy = reader.ReadInt16();
	        var gyroz = reader.ReadInt16();
	

		return new Ns2proInput
		{
	            Buttons = buttons,
	            Lx = lx,
	            Ly = ly,
	            Rx = rx,
	            Ry = ry,
	            Accelx = accelx,
	            Accely = accely,
	            Accelz = accelz,
	            Gyrox = gyrox,
	            Gyroy = gyroy,
	            Gyroz = gyroz,
	        };
	    }
}
