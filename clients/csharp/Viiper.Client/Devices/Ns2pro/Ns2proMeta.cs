// Auto-generated VIIPER C# Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

using System;
using System.Collections.Generic;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Viiper.Client.Devices.Ns2pro;


public class MetaState
{
    [JsonPropertyName("serial_number")]
    public string SerialNumber { get; set; } = string.Empty;

    [JsonPropertyName("battery_level")]
    public byte BatteryLevel { get; set; }

    [JsonPropertyName("charging")]
    public bool Charging { get; set; }

    [JsonPropertyName("external_power")]
    public bool ExternalPower { get; set; }

    [JsonPropertyName("battery_volts")]
    public ushort BatteryVolts { get; set; }


    public Dictionary<string, object?> ToMap()
    {
        var json = JsonSerializer.Serialize(this);
        return JsonSerializer.Deserialize<Dictionary<string, object?>>(json) ?? new Dictionary<string, object?>();
    }

    public static MetaState FromMap(Dictionary<string, object?> map)
    {
        var json = JsonSerializer.Serialize(map);
        return JsonSerializer.Deserialize<MetaState>(json) ?? new MetaState();
    }
}


