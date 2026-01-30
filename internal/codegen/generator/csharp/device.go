package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const deviceTemplate = `{{writeFileHeader}}using System.IO;
using System.Net.Sockets;
using System.Threading.Channels;

namespace Viiper.Client;

/// <summary>
/// Interface for binary serializable input payloads sent to a VIIPER device stream.
/// </summary>
public interface IBinarySerializable
{
	void Write(BinaryWriter writer);
}

/// <summary>
/// Represents a live device stream connection allowing input sending and receiving raw output bytes.
/// </summary>
public sealed class ViiperDevice : IAsyncDisposable, IDisposable
{
	private readonly TcpClient _client;
	private readonly Stream _stream;
	private readonly CancellationTokenSource _cts = new();
	private Task? _readLoop;
	private bool _disposed;
	private Func<Stream, Task>? _onOutput;
	private Action? _onDisconnect;

	/// <summary>
	/// Callback invoked when output data is available from the device.
	/// The callback receives the stream and must read the exact number of bytes expected.
	/// </summary>
	public Func<Stream, Task>? OnOutput
	{
		get => _onOutput;
		set
		{
			_onOutput = value;
			if (_onOutput != null && _readLoop == null)
			{
				_readLoop = Task.Run(ReadLoopAsync);
			}
		}
	}

	public Action? OnDisconnect
	{
		get => _onDisconnect;
		set => _onDisconnect = value;
	}

	internal ViiperDevice(TcpClient client, Stream stream)
	{
		_client = client;
		_stream = stream;
	}

	/// <summary>
	/// Send a binary-serializable input payload to the device.
	/// </summary>
	public async Task SendAsync<T>(T payload, CancellationToken cancellationToken = default) where T : IBinarySerializable
	{
		ThrowIfDisposed();
		using var ms = new MemoryStream();
		using (var bw = new BinaryWriter(ms, System.Text.Encoding.UTF8, leaveOpen: true))
		{
			payload.Write(bw);
		}
		var buf = ms.ToArray();
		await _stream.WriteAsync(buf, 0, buf.Length, cancellationToken);
	}

	/// <summary>
	/// Send raw bytes to the device (advanced usage).
	/// </summary>
	public async Task SendRawAsync(byte[] data, CancellationToken cancellationToken = default)
	{
		ThrowIfDisposed();
		await _stream.WriteAsync(data, 0, data.Length, cancellationToken);
	}

	private async Task ReadLoopAsync()
	{
		try
		{
			while (!_cts.IsCancellationRequested && _onOutput != null)
			{
				await _onOutput(_stream).ConfigureAwait(false);
			}
		}
		catch (OperationCanceledException)
		{
			// normal during shutdown
		}
		catch (Exception)
		{
			// swallow; user can detect via absence of further events
		}
		_onDisconnect?.Invoke();
	}

	private void ThrowIfDisposed()
	{
		if (_disposed)
			throw new ObjectDisposedException(nameof(ViiperDevice));
	}

	/// <summary>
	/// Dispose synchronously.
	/// </summary>
	public void Dispose()
	{
		if (_disposed) return;
		_disposed = true;
		_cts.Cancel();
		try { _readLoop?.Wait(); } catch { }
		_stream.Dispose();
		_client.Dispose();
		_cts.Dispose();
		GC.SuppressFinalize(this);
	}

	/// <summary>
	/// Dispose asynchronously awaiting read loop completion.
	/// </summary>
	public async ValueTask DisposeAsync()
	{
		if (_disposed) return;
		_disposed = true;
		_cts.Cancel();
		if (_readLoop != null)
			try { await _readLoop.ConfigureAwait(false); } catch { }
		_stream.Dispose();
		_client.Dispose();
		_cts.Dispose();
		GC.SuppressFinalize(this);
	}
}
`

func generateDevice(logger *slog.Logger, projectDir string, md *meta.Metadata) error {
	logger.Debug("Generating ViiperDevice stream wrapper")
	outputFile := filepath.Join(projectDir, "ViiperDevice.cs")
	tmpl := template.Must(template.New("device").Funcs(template.FuncMap{
		"writeFileHeader": writeFileHeader,
	}).Parse(deviceTemplate))
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create ViiperDevice.cs: %w", err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, md); err != nil {
		return fmt.Errorf("execute device template: %w", err)
	}
	logger.Info("Generated ViiperDevice", "file", outputFile)
	return nil
}
