// Auto-generated VIIPER C# Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

using System;
using System.Collections.Generic;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Viiper.Client.Devices.Dualshock4;


public class MetaState
{
    [JsonPropertyName("serial_number")]
    public string SerialNumber { get; set; } = string.Empty;

    [JsonPropertyName("board")]
    public string Board { get; set; } = string.Empty;

    [JsonPropertyName("build_time")]
    public DateTime BuildTime { get; set; }

    [JsonPropertyName("battery_status")]
    public byte BatteryStatus { get; set; }

    [JsonPropertyName("temperature_celsius")]
    public double TemperatureCelsius { get; set; }

    [JsonPropertyName("battery_voltage")]
    public double BatteryVoltage { get; set; }


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


