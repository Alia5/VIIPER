using System;
using System.IO;

namespace Viiper.Client.Devices.Xbox360;

/// <summary>
/// Wire protocol Output message for Xbox360 device.
/// </summary>
public class Xbox360Output : IBinarySerializable
{
    public required byte Left { get; set; }
    public required byte Right { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Left);
        writer.Write(Right);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static Xbox360Output Read(BinaryReader reader)
    {
	        var left = reader.ReadByte();
	        var right = reader.ReadByte();
	

		return new Xbox360Output
		{
	            Left = left,
	            Right = right,
	        };
	    }
}
