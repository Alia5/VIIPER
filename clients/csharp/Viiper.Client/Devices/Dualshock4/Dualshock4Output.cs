using System;
using System.IO;

namespace Viiper.Client.Devices.Dualshock4;

/// <summary>
/// Wire protocol Output message for Dualshock4 device.
/// </summary>
public class Dualshock4Output : IBinarySerializable
{
    public required byte Updateflags { get; set; }
    public required byte Rumblesmall { get; set; }
    public required byte Rumblelarge { get; set; }
    public required byte Ledred { get; set; }
    public required byte Ledgreen { get; set; }
    public required byte Ledblue { get; set; }
    public required byte Flashon { get; set; }
    public required byte Flashoff { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Updateflags);
        writer.Write(Rumblesmall);
        writer.Write(Rumblelarge);
        writer.Write(Ledred);
        writer.Write(Ledgreen);
        writer.Write(Ledblue);
        writer.Write(Flashon);
        writer.Write(Flashoff);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static Dualshock4Output Read(BinaryReader reader)
    {
	        var updateflags = reader.ReadByte();
	        var rumblesmall = reader.ReadByte();
	        var rumblelarge = reader.ReadByte();
	        var ledred = reader.ReadByte();
	        var ledgreen = reader.ReadByte();
	        var ledblue = reader.ReadByte();
	        var flashon = reader.ReadByte();
	        var flashoff = reader.ReadByte();
	

		return new Dualshock4Output
		{
	            Updateflags = updateflags,
	            Rumblesmall = rumblesmall,
	            Rumblelarge = rumblelarge,
	            Ledred = ledred,
	            Ledgreen = ledgreen,
	            Ledblue = ledblue,
	            Flashon = flashon,
	            Flashoff = flashoff,
	        };
	    }
}
