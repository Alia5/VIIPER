// Auto-generated VIIPER C# Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Viiper.Client.Types;

namespace Viiper.Client;

/// <summary>
/// VIIPER management API client for bus and device control
/// </summary>
public class ViiperClient : IDisposable
{
    private readonly string _host;
    private readonly int _port;
    private readonly string _password;
    private bool _disposed;

    /// <summary>
    /// Creates a new VIIPER client instance
    /// </summary>
    /// <param name="host">VIIPER server hostname or IP address</param>
    /// <param name="port">VIIPER API server port (default: 3242)</param>
    /// <param name="password">Authentication password (default: "" = no auth). Empty string explicitly means no authentication.</param>
    public ViiperClient(string host, int port = 3242, string password = "")
    {
        _host = host ?? throw new ArgumentNullException(nameof(host));
        _port = port;
        _password = password ?? "";
    }

    /// <summary>
    /// Ping: ping
    /// </summary>
    /// <returns>PingResponse</returns>
    public async Task<PingResponse> PingAsync(CancellationToken cancellationToken = default)
    {
        var path = "ping";
        
		string? payload = null;
        return await SendRequestAsync<PingResponse>(path, payload, cancellationToken);
    }

    /// <summary>
    /// BusList: bus/list
    /// </summary>
    /// <returns>BusListResponse</returns>
    public async Task<BusListResponse> BusListAsync(CancellationToken cancellationToken = default)
    {
        var path = "bus/list";
        
		string? payload = null;
        return await SendRequestAsync<BusListResponse>(path, payload, cancellationToken);
    }

    /// <summary>
    /// BusCreate: bus/create
    /// </summary>
    /// <returns>BusCreateResponse</returns>
    public async Task<BusCreateResponse> BusCreateAsync(uint? id, CancellationToken cancellationToken = default)
    {
        var path = "bus/create";
        
		string? payload = id?.ToString();
        return await SendRequestAsync<BusCreateResponse>(path, payload, cancellationToken);
    }

    /// <summary>
    /// BusRemove: bus/remove
    /// </summary>
    /// <returns>BusRemoveResponse</returns>
    public async Task<BusRemoveResponse> BusRemoveAsync(uint id, CancellationToken cancellationToken = default)
    {
        var path = "bus/remove";
        
		string? payload = id.ToString();
        return await SendRequestAsync<BusRemoveResponse>(path, payload, cancellationToken);
    }

    /// <summary>
    /// BusDevicesList: bus/{id}/list
    /// </summary>
    /// <returns>DevicesListResponse</returns>
    public async Task<DevicesListResponse> BusDevicesListAsync(uint id, CancellationToken cancellationToken = default)
    {
        var path = "bus/{id}/list".Replace("{id}", id.ToString());
        
		string? payload = null;
        return await SendRequestAsync<DevicesListResponse>(path, payload, cancellationToken);
    }

    /// <summary>
    /// BusDeviceAdd: bus/{id}/add
    /// </summary>
    /// <returns>Device</returns>
    public async Task<Device> BusDeviceAddAsync(uint id, DeviceCreateRequest devicecreaterequest, CancellationToken cancellationToken = default)
    {
        var path = "bus/{id}/add".Replace("{id}", id.ToString());
        
		string? payload = JsonSerializer.Serialize(devicecreaterequest);
        return await SendRequestAsync<Device>(path, payload, cancellationToken);
    }

    /// <summary>
    /// BusDeviceRemove: bus/{id}/remove
    /// </summary>
    /// <returns>DeviceRemoveResponse</returns>
    public async Task<DeviceRemoveResponse> BusDeviceRemoveAsync(uint id, string value, CancellationToken cancellationToken = default)
    {
        var path = "bus/{id}/remove".Replace("{id}", id.ToString());
        
		string? payload = value;
        return await SendRequestAsync<DeviceRemoveResponse>(path, payload, cancellationToken);
    }

    private async Task<T> SendRequestAsync<T>(string path, string? payload, CancellationToken cancellationToken)
    {
        using var client = new TcpClient();
        await client.ConnectAsync(_host, _port, cancellationToken);
        client.NoDelay = true;
        
        Stream stream = client.GetStream();
        
        if (!string.IsNullOrEmpty(_password))
        {
            stream = await ViiperAuth.PerformHandshakeAsync(stream, _password, cancellationToken);
        }
        
        string commandLine = path.ToLowerInvariant();
        if (!string.IsNullOrEmpty(payload))
        {
            commandLine += " " + payload;
        }
		commandLine += "\0";
        
        var requestBytes = Encoding.UTF8.GetBytes(commandLine);
        await stream.WriteAsync(requestBytes, cancellationToken);
        
        var responseBuilder = new StringBuilder();
        var buffer = new byte[128];
        int bytesRead;
        
        while ((bytesRead = await stream.ReadAsync(buffer, cancellationToken)) > 0)
        {
            responseBuilder.Append(Encoding.UTF8.GetString(buffer, 0, bytesRead));
        }
        
		var responseJson = responseBuilder.ToString().TrimEnd('\n');
		// Typed error detection (RFC 7807 style): check for status field prefix
		if (responseJson.StartsWith("{\"status\":"))
		{
			throw new InvalidOperationException($"VIIPER error response: {responseJson}");
		}
		var response = JsonSerializer.Deserialize<T>(responseJson)
			?? throw new InvalidOperationException("Failed to deserialize response");
        
        return response;
    }

    /// <summary>
    /// Creates a device stream connection for sending input and receiving output
    /// </summary>
    /// <param name="busId">Bus ID</param>
    /// <param name="devId">Device ID</param>
    /// <param name="cancellationToken">Cancellation token</param>
    /// <returns>ViiperDevice stream wrapper</returns>
	public async Task<ViiperDevice> ConnectDeviceAsync(uint busId, string devId, CancellationToken cancellationToken = default)
	{
		var client = new TcpClient();
		await client.ConnectAsync(_host, _port, cancellationToken);
		client.NoDelay = true;
		Stream stream = client.GetStream();
		
		if (!string.IsNullOrEmpty(_password))
		{
			stream = await ViiperAuth.PerformHandshakeAsync(stream, _password, cancellationToken);
		}
		
		// Streaming handshake uses null terminator (same framing as management).
		var streamPath = $"bus/{busId}/{devId}\0";
		var handshake = Encoding.UTF8.GetBytes(streamPath);
		await stream.WriteAsync(handshake, cancellationToken);
		return new ViiperDevice(client, stream);
	}

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        GC.SuppressFinalize(this);
    }
}
