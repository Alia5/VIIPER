using System;
using System.IO;

namespace Viiper.Client.Devices.Ns2pro;

/// <summary>
/// Wire protocol Output message for Ns2pro device.
/// </summary>
public class Ns2proOutput : IBinarySerializable
{
    public byte[] Leftrumble { get; set; } = new byte[16];
    public byte[] Rightrumble { get; set; } = new byte[16];
    public required byte Flags { get; set; }
    public required byte Playerledmask { get; set; }

    public void Write(BinaryWriter writer)
    {
        for (int i = 0; i < 16; i++)
		{
			writer.Write((Leftrumble != null && i < Leftrumble.Length) ? Leftrumble[i] : default(byte));
		}
        for (int i = 0; i < 16; i++)
		{
			writer.Write((Rightrumble != null && i < Rightrumble.Length) ? Rightrumble[i] : default(byte));
		}
        writer.Write(Flags);
        writer.Write(Playerledmask);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static Ns2proOutput Read(BinaryReader reader)
    {
	        var leftrumble = new byte[16];
		for (int i = 0; i < 16; i++)
		{
		    leftrumble[i] = reader.ReadByte();
		}
	        var rightrumble = new byte[16];
		for (int i = 0; i < 16; i++)
		{
		    rightrumble[i] = reader.ReadByte();
		}
	        var flags = reader.ReadByte();
	        var playerledmask = reader.ReadByte();
	

		return new Ns2proOutput
		{
	            Leftrumble = leftrumble,
	            Rightrumble = rightrumble,
	            Flags = flags,
	            Playerledmask = playerledmask,
	        };
	    }
}
