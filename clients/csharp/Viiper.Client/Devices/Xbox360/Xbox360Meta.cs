// Auto-generated VIIPER C# Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

using System;
using System.Collections.Generic;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Viiper.Client.Devices.Xbox360;


public class Xbox360CreateOptions
{
    [JsonPropertyName("subType")]
    public byte? SubType { get; set; }


    public Dictionary<string, object?> ToMap()
    {
        var json = JsonSerializer.Serialize(this);
        return JsonSerializer.Deserialize<Dictionary<string, object?>>(json) ?? new Dictionary<string, object?>();
    }

    public static Xbox360CreateOptions FromMap(Dictionary<string, object?> map)
    {
        var json = JsonSerializer.Serialize(map);
        return JsonSerializer.Deserialize<Xbox360CreateOptions>(json) ?? new Xbox360CreateOptions();
    }
}


