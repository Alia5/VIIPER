using System;
using System.IO;

namespace Viiper.Client.Devices.Dualsense;

/// <summary>
/// Wire protocol Output message for Dualsense device.
/// </summary>
public class DualsenseOutput : IBinarySerializable
{
    public required byte Updateflags1 { get; set; }
    public required byte Updateflags2 { get; set; }
    public required byte Updateflags3 { get; set; }
    public required byte Rumblesmall { get; set; }
    public required byte Rumblelarge { get; set; }
    public required byte Ledred { get; set; }
    public required byte Ledgreen { get; set; }
    public required byte Ledblue { get; set; }
    public required byte Playerleds { get; set; }

    public void Write(BinaryWriter writer)
    {
        writer.Write(Updateflags1);
        writer.Write(Updateflags2);
        writer.Write(Updateflags3);
        writer.Write(Rumblesmall);
        writer.Write(Rumblelarge);
        writer.Write(Ledred);
        writer.Write(Ledgreen);
        writer.Write(Ledblue);
        writer.Write(Playerleds);
    }

    /// <summary>
    /// Read from binary stream (for receiving output from server).
    /// </summary>
    public static DualsenseOutput Read(BinaryReader reader)
    {
	        var updateflags1 = reader.ReadByte();
	        var updateflags2 = reader.ReadByte();
	        var updateflags3 = reader.ReadByte();
	        var rumblesmall = reader.ReadByte();
	        var rumblelarge = reader.ReadByte();
	        var ledred = reader.ReadByte();
	        var ledgreen = reader.ReadByte();
	        var ledblue = reader.ReadByte();
	        var playerleds = reader.ReadByte();
	

		return new DualsenseOutput
		{
	            Updateflags1 = updateflags1,
	            Updateflags2 = updateflags2,
	            Updateflags3 = updateflags3,
	            Rumblesmall = rumblesmall,
	            Rumblelarge = rumblelarge,
	            Ledred = ledred,
	            Ledgreen = ledgreen,
	            Ledblue = ledblue,
	            Playerleds = playerleds,
	        };
	    }
}
