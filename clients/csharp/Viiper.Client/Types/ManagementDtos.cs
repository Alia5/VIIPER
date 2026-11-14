// Auto-generated VIIPER C# Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace Viiper.Client.Types;


/// <summary>
/// APIError DTO
/// </summary>
public class APIError
{
    [JsonPropertyName("status")]
    public required int Status { get; set; }

    [JsonPropertyName("title")]
    public required string Title { get; set; }

    [JsonPropertyName("detail")]
    public required string Detail { get; set; }

}


/// <summary>
/// PingResponse DTO
/// </summary>
public class PingResponse
{
    [JsonPropertyName("server")]
    public required string Server { get; set; }

    [JsonPropertyName("version")]
    public required string Version { get; set; }

}


/// <summary>
/// BusListResponse DTO
/// </summary>
public class BusListResponse
{
    [JsonPropertyName("buses")]
    public required uint[] Buses { get; set; }

}


/// <summary>
/// BusCreateResponse DTO
/// </summary>
public class BusCreateResponse
{
    [JsonPropertyName("busId")]
    public required uint BusID { get; set; }

}


/// <summary>
/// BusRemoveResponse DTO
/// </summary>
public class BusRemoveResponse
{
    [JsonPropertyName("busId")]
    public required uint BusID { get; set; }

}


/// <summary>
/// Device DTO
/// </summary>
public class Device
{
    [JsonPropertyName("busId")]
    public required uint BusID { get; set; }

    [JsonPropertyName("devId")]
    public required string DevID { get; set; }

    [JsonPropertyName("vid")]
    public required string Vid { get; set; }

    [JsonPropertyName("pid")]
    public required string Pid { get; set; }

    [JsonPropertyName("type")]
    public required string Type { get; set; }

    [JsonPropertyName("deviceSpecific")]
    public required Dictionary<string, object?> DeviceSpecific { get; set; }

}


/// <summary>
/// DevicesListResponse DTO
/// </summary>
public class DevicesListResponse
{
    [JsonPropertyName("devices")]
    public required Device[] Devices { get; set; }

}


/// <summary>
/// DeviceRemoveResponse DTO
/// </summary>
public class DeviceRemoveResponse
{
    [JsonPropertyName("busId")]
    public required uint BusID { get; set; }

    [JsonPropertyName("devId")]
    public required string DevID { get; set; }

}


/// <summary>
/// DeviceCreateRequest DTO
/// </summary>
public class DeviceCreateRequest
{
    [JsonPropertyName("type")]
    public string? Type { get; set; }

    [JsonPropertyName("idVendor")]
    public ushort? IDVendor { get; set; }

    [JsonPropertyName("idProduct")]
    public ushort? IDProduct { get; set; }

    [JsonPropertyName("deviceSpecific")]
    public Dictionary<string, object?>? DeviceSpecific { get; set; }

}


