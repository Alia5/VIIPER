using System;
using System.IO;

namespace Viiper.Client.Devices.Dualsense;

/// <summary>
/// Wire protocol Input message for Dualsense device.
/// </summary>
public class DualsenseInput : IBinarySerializable
{
    public required sbyte Sticklx { get; set; }
    public required sbyte Stickly { get; set; }
    public required sbyte Stickrx { get; set; }
    public required sbyte Stickry { get; set; }
    public required uint Buttons { get; set; }
    public required byte Dpad { get; set; }
    public required byte Triggerl2 { get; set; }
    public required byte Triggerr2 { get; set; }
    public required ushort Touch1x { get; set; }
    public required ushort Touch1y { get; set; }
    public required byte Touch1active { get; set; }
    public required ushort Touch2x { get; set; }
    public required ushort Touch2y { get; set; }
    public required byte Touch2active { get; set; }
    public required short Gyrox { get; set; }
    public required short Gyroy { get; set; }
    public required short Gyroz { get; set; }
    public required short Accelx { get; set; }
    public required short Accely { get; set; }
    public required short Accelz { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Sticklx);
        writer.Write(Stickly);
        writer.Write(Stickrx);
        writer.Write(Stickry);
        writer.Write(Buttons);
        writer.Write(Dpad);
        writer.Write(Triggerl2);
        writer.Write(Triggerr2);
        writer.Write(Touch1x);
        writer.Write(Touch1y);
        writer.Write(Touch1active);
        writer.Write(Touch2x);
        writer.Write(Touch2y);
        writer.Write(Touch2active);
        writer.Write(Gyrox);
        writer.Write(Gyroy);
        writer.Write(Gyroz);
        writer.Write(Accelx);
        writer.Write(Accely);
        writer.Write(Accelz);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static DualsenseInput Read(BinaryReader reader)
    {
	        var sticklx = reader.ReadSByte();
	        var stickly = reader.ReadSByte();
	        var stickrx = reader.ReadSByte();
	        var stickry = reader.ReadSByte();
	        var buttons = reader.ReadUInt32();
	        var dpad = reader.ReadByte();
	        var triggerl2 = reader.ReadByte();
	        var triggerr2 = reader.ReadByte();
	        var touch1x = reader.ReadUInt16();
	        var touch1y = reader.ReadUInt16();
	        var touch1active = reader.ReadByte();
	        var touch2x = reader.ReadUInt16();
	        var touch2y = reader.ReadUInt16();
	        var touch2active = reader.ReadByte();
	        var gyrox = reader.ReadInt16();
	        var gyroy = reader.ReadInt16();
	        var gyroz = reader.ReadInt16();
	        var accelx = reader.ReadInt16();
	        var accely = reader.ReadInt16();
	        var accelz = reader.ReadInt16();
	

		return new DualsenseInput
		{
	            Sticklx = sticklx,
	            Stickly = stickly,
	            Stickrx = stickrx,
	            Stickry = stickry,
	            Buttons = buttons,
	            Dpad = dpad,
	            Triggerl2 = triggerl2,
	            Triggerr2 = triggerr2,
	            Touch1x = touch1x,
	            Touch1y = touch1y,
	            Touch1active = touch1active,
	            Touch2x = touch2x,
	            Touch2y = touch2y,
	            Touch2active = touch2active,
	            Gyrox = gyrox,
	            Gyroy = gyroy,
	            Gyroz = gyroz,
	            Accelx = accelx,
	            Accely = accely,
	            Accelz = accelz,
	        };
	    }
}
