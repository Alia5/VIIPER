using System;
using System.IO;

namespace Viiper.Client.Devices.Mouse;

/// <summary>
/// Wire protocol Input message for Mouse device.
/// </summary>
public class MouseInput : IBinarySerializable
{
    public required byte Buttons { get; set; }
    public required short Dx { get; set; }
    public required short Dy { get; set; }
    public required short Wheel { get; set; }
    public required short Pan { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Buttons);
        writer.Write(Dx);
        writer.Write(Dy);
        writer.Write(Wheel);
        writer.Write(Pan);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static MouseInput Read(BinaryReader reader)
    {
	        var buttons = reader.ReadByte();
	        var dx = reader.ReadInt16();
	        var dy = reader.ReadInt16();
	        var wheel = reader.ReadInt16();
	        var pan = reader.ReadInt16();
	

		return new MouseInput
		{
	            Buttons = buttons,
	            Dx = dx,
	            Dy = dy,
	            Wheel = wheel,
	            Pan = pan,
	        };
	    }
}
