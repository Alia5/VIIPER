using System;
using System.IO;

namespace Viiper.Client.Devices.Keyboard;

/// <summary>
/// Wire protocol Input message for Keyboard device.
/// </summary>
public class KeyboardInput : IBinarySerializable
{
    public required byte Modifiers { get; set; }
    public required byte Count { get; set; }
    public required byte[] Keys { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Modifiers);
        writer.Write(Count);
        for (int i = 0; i < Count; i++)
		{
			writer.Write(Keys[i]);
		}
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static KeyboardInput Read(BinaryReader reader)
    {
	        var modifiers = reader.ReadByte();
	        var count = reader.ReadByte();
	        var keys = new byte[count];
		for (int i = 0; i < count; i++)
		{
		    keys[i] = reader.ReadByte();
		}
	

		return new KeyboardInput
		{
	            Modifiers = modifiers,
	            Count = count,
	            Keys = keys,
	        };
	    }
}
