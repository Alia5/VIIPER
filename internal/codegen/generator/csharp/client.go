package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const clientTemplate = `{{writeFileHeader}}using System.Net.Sockets;
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
{{range .Routes}}{{if eq .Method "Register"}}
    /// <summary>
    /// {{.Handler}}: {{.Path}}
    /// </summary>{{if .ResponseDTO}}
    /// <returns>{{.ResponseDTO}}</returns>{{end}}
    public async Task<{{if .ResponseDTO}}{{.ResponseDTO}}{{else}}bool{{end}}> {{.Handler}}Async({{generateMethodParams .}}CancellationToken cancellationToken = default)
    {
        var path = "{{.Path}}"{{range $key, $value := .PathParams}}.Replace("{{lb}}{{$key}}{{rb}}", {{toCamelCase $key}}.ToString()){{end}};
        {{/* Build payload based on classification */}}
		{{if eq .Payload.Kind "none"}}string? payload = null;{{else if eq .Payload.Kind "json"}}string? payload = JsonSerializer.Serialize({{payloadParamNameCS .}});{{else if eq .Payload.Kind "numeric"}}{{if .Payload.Required}}string? payload = {{payloadParamNameCS .}}.ToString();{{else}}string? payload = {{payloadParamNameCS .}}?.ToString();{{end}}{{else if eq .Payload.Kind "string"}}string? payload = {{payloadParamNameCS .}};{{end}}
        {{if .ResponseDTO}}return await SendRequestAsync<{{.ResponseDTO}}>(path, payload, cancellationToken);{{else}}await SendRequestAsync<object>(path, payload, cancellationToken);
        return true;{{end}}
    }
{{end}}{{end}}
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
		var streamPath = $"bus/{{lb}}busId{{rb}}/{{lb}}devId{{rb}}\0";
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
`

func generateClient(logger *slog.Logger, projectDir string, md *meta.Metadata) error {
	logger.Debug("Generating ViiperClient management API")

	outputFile := filepath.Join(projectDir, "ViiperClient.cs")

	funcMap := template.FuncMap{
		"toCamelCase":          toCamelCase,
		"writeFileHeader":      writeFileHeader,
		"generateMethodParams": generateMethodParams,
		"payloadParamNameCS":   payloadParamNameCS,
		"lb":                   func() string { return "{" },
		"rb":                   func() string { return "}" },
	}

	tmpl, err := template.New("client").Funcs(funcMap).Parse(clientTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	data := struct {
		Routes []scanner.RouteInfo
	}{
		Routes: md.Routes,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	logger.Info("Generated ViiperClient", "file", outputFile)
	return nil
}

func generateMethodParams(route scanner.RouteInfo) string {
	var params []string
	for key := range route.PathParams {
		params = append(params, fmt.Sprintf("uint %s", toCamelCase(key)))
	}
	switch route.Payload.Kind {
	case scanner.PayloadJSON:
		name := payloadParamNameCS(route)
		typeName := route.Payload.RawType
		if typeName == "" {
			typeName = "object"
		}
		params = append(params, fmt.Sprintf("%s %s", typeName, name))
	case scanner.PayloadNumeric:
		name := payloadParamNameCS(route)
		typeName := "uint"
		if strings.HasPrefix(route.Payload.RawType, "int") && !strings.HasPrefix(route.Payload.RawType, "uint") {
			typeName = "int"
		} else if strings.HasPrefix(route.Payload.RawType, "uint") {
			typeName = "uint"
		}
		if !route.Payload.Required {
			typeName += "?"
		}
		params = append(params, fmt.Sprintf("%s %s", typeName, name))
	case scanner.PayloadString:
		name := payloadParamNameCS(route)
		typeName := "string"
		if !route.Payload.Required {
			typeName += "?"
		}
		params = append(params, fmt.Sprintf("%s %s", typeName, name))
	}
	if len(params) == 0 {
		return ""
	}
	return strings.Join(params, ", ") + ", "
}

func payloadParamNameCS(route scanner.RouteInfo) string {
	if route.Payload.Kind == scanner.PayloadNone {
		return ""
	}
	hint := route.Payload.ParserHint
	if hint == "" {
		return "payload"
	}
	switch route.Payload.Kind {
	case scanner.PayloadNumeric:
		if strings.Contains(strings.ToLower(hint), "id") || strings.HasPrefix(hint, "uint") || strings.HasPrefix(hint, "int") {
			return "id"
		}
		return "value"
	case scanner.PayloadJSON:
		if route.Payload.RawType != "" {
			return toCamelCase(route.Payload.RawType)
		}
		return "request"
	case scanner.PayloadString:
		return "value"
	}
	return "payload"
}
