using System;
using System.IO;

namespace Viiper.Client.Devices.Keyboard;

/// <summary>
/// Wire protocol Output message for Keyboard device.
/// </summary>
public class KeyboardOutput : IBinarySerializable
{
    public required byte Leds { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Leds);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static KeyboardOutput Read(BinaryReader reader)
    {
	        var leds = reader.ReadByte();
	

		return new KeyboardOutput
		{
	            Leds = leds,
	        };
	    }
}
