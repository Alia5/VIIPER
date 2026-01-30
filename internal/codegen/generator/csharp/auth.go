package csharp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const authHelperTemplate = `// Auto-generated VIIPER C# Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

using System;
using System.IO;
using System.Net.Sockets;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

namespace Viiper.Client;

/// <summary>
/// Authentication and encryption utilities for VIIPER client
/// </summary>
internal static class ViiperAuth
{
    private const string HandshakeMagic = "eVI1\0";
    private const int NonceSize = 32;
    private const string AuthContext = "VIIPER-Auth-v1";
    private const string SessionContext = "VIIPER-Session-v1";
    private const int PBKDF2Iterations = 100000;
    private const string PBKDF2Salt = "VIIPER-Key-v1";

    /// <summary>
    /// Derive a 32-byte key from password using PBKDF2-SHA256
    /// </summary>
    public static byte[] DeriveKey(string password)
    {
        if (string.IsNullOrEmpty(password))
        {
            throw new ArgumentException("Password cannot be empty", nameof(password));
        }

        using var pbkdf2 = new Rfc2898DeriveBytes(
            password,
            Encoding.UTF8.GetBytes(PBKDF2Salt),
            PBKDF2Iterations,
            HashAlgorithmName.SHA256
        );
        return pbkdf2.GetBytes(32);
    }

    /// <summary>
    /// Derive session key from key and nonces using SHA-256
    /// </summary>
    public static byte[] DeriveSessionKey(byte[] key, byte[] serverNonce, byte[] clientNonce)
    {
        using var sha256 = SHA256.Create();
        var combined = new byte[key.Length + serverNonce.Length + clientNonce.Length + SessionContext.Length];
        var offset = 0;
        
        Buffer.BlockCopy(key, 0, combined, offset, key.Length);
        offset += key.Length;
        
        Buffer.BlockCopy(serverNonce, 0, combined, offset, serverNonce.Length);
        offset += serverNonce.Length;
        
        Buffer.BlockCopy(clientNonce, 0, combined, offset, clientNonce.Length);
        offset += clientNonce.Length;
        
        var contextBytes = Encoding.UTF8.GetBytes(SessionContext);
        Buffer.BlockCopy(contextBytes, 0, combined, offset, contextBytes.Length);
        
        return sha256.ComputeHash(combined);
    }

    /// <summary>
    /// Perform authentication handshake with VIIPER server
    /// Returns a stream (potentially wrapped with encryption) if handshake succeeds
    /// </summary>
    public static async Task<Stream> PerformHandshakeAsync(
        Stream stream, 
        string password, 
        CancellationToken cancellationToken = default)
    {
        var key = DeriveKey(password);
        
        var clientNonce = new byte[NonceSize];
        using (var rng = RandomNumberGenerator.Create())
        {
            rng.GetBytes(clientNonce);
        }
        
        byte[] authTag;
        using (var hmac = new HMACSHA256(key))
        {
            var authData = new byte[AuthContext.Length + clientNonce.Length];
            var authContextBytes = Encoding.UTF8.GetBytes(AuthContext);
            Buffer.BlockCopy(authContextBytes, 0, authData, 0, authContextBytes.Length);
            Buffer.BlockCopy(clientNonce, 0, authData, authContextBytes.Length, clientNonce.Length);
            authTag = hmac.ComputeHash(authData);
        }
        
        var handshakeMsg = new byte[HandshakeMagic.Length + clientNonce.Length + authTag.Length];
        var magicBytes = Encoding.UTF8.GetBytes(HandshakeMagic);
        Buffer.BlockCopy(magicBytes, 0, handshakeMsg, 0, magicBytes.Length);
        Buffer.BlockCopy(clientNonce, 0, handshakeMsg, magicBytes.Length, clientNonce.Length);
        Buffer.BlockCopy(authTag, 0, handshakeMsg, magicBytes.Length + clientNonce.Length, authTag.Length);
        
        await stream.WriteAsync(handshakeMsg, 0, handshakeMsg.Length, cancellationToken);
        
        var response = new byte[3 + NonceSize];
        await ReadExactlyAsync(stream, response, cancellationToken);
        
        var prefix = Encoding.UTF8.GetString(response, 0, 3);
        if (prefix != "OK\0")
        {
            var errorBuilder = new StringBuilder();
            errorBuilder.Append(Encoding.UTF8.GetString(response));
            
            var buffer = new byte[4096];
            int bytesRead;
            while ((bytesRead = await stream.ReadAsync(buffer, 0, buffer.Length, cancellationToken)) > 0)
            {
                errorBuilder.Append(Encoding.UTF8.GetString(buffer, 0, bytesRead));
            }
            
            var errorResponse = errorBuilder.ToString().Trim();
            try
            {
                var error = JsonSerializer.Deserialize<JsonDocument>(errorResponse);
                if (error != null && error.RootElement.TryGetProperty("title", out var title))
                {
                    throw new InvalidOperationException($"Authentication failed: {title.GetString()}");
                }
            }
            catch (JsonException)
            {
                // Not JSON, throw raw response
            }
            throw new InvalidOperationException($"Invalid handshake response: {errorResponse}");
        }
        
        var serverNonce = new byte[NonceSize];
        Buffer.BlockCopy(response, 3, serverNonce, 0, NonceSize);
        
        var sessionKey = DeriveSessionKey(key, serverNonce, clientNonce);
        
        return new EncryptedStream(stream, sessionKey);
    }

    /// <summary>
    /// Read exactly the specified number of bytes from stream
    /// </summary>
    private static async Task ReadExactlyAsync(Stream stream, byte[] buffer, CancellationToken cancellationToken)
    {
        int totalRead = 0;
        while (totalRead < buffer.Length)
        {
            int bytesRead = await stream.ReadAsync(
                buffer, 
                totalRead, 
                buffer.Length - totalRead, 
                cancellationToken
            );
            
            if (bytesRead == 0)
            {
                throw new EndOfStreamException("Connection closed before receiving full response");
            }
            
            totalRead += bytesRead;
        }
    }
}

/// <summary>
/// Encrypted stream wrapper using ChaCha20-Poly1305
/// Requires .NET 5+ for ChaCha20Poly1305 support
/// </summary>
internal class EncryptedStream : Stream
{
    private readonly Stream _innerStream;
    private readonly byte[] _sessionKey;
    private ulong _sendCounter = 0;
    private byte[] _recvBuffer = Array.Empty<byte>();
    private int _recvBufferPos = 0;

    public EncryptedStream(Stream innerStream, byte[] sessionKey)
    {
        _innerStream = innerStream ?? throw new ArgumentNullException(nameof(innerStream));
        _sessionKey = sessionKey ?? throw new ArgumentNullException(nameof(sessionKey));
        
        if (sessionKey.Length != 32)
        {
            throw new ArgumentException("Session key must be 32 bytes", nameof(sessionKey));
        }
    }

    public override bool CanRead => _innerStream.CanRead;
    public override bool CanSeek => false;
    public override bool CanWrite => _innerStream.CanWrite;
    public override long Length => throw new NotSupportedException();
    public override long Position 
    { 
        get => throw new NotSupportedException(); 
        set => throw new NotSupportedException(); 
    }

    public override void Write(byte[] buffer, int offset, int count)
    {
        WriteAsync(buffer, offset, count).GetAwaiter().GetResult();
    }

    public override async Task WriteAsync(byte[] buffer, int offset, int count, CancellationToken cancellationToken)
    {
        var nonce = new byte[12];
        var counterBytes = BitConverter.GetBytes(_sendCounter);
        if (BitConverter.IsLittleEndian)
        {
            Array.Reverse(counterBytes); // Convert to big-endian
        }
        Buffer.BlockCopy(counterBytes, 0, nonce, 4, 8);
        _sendCounter++;

        // Encrypt with ChaCha20-Poly1305
        var plaintext = new byte[count];
        Buffer.BlockCopy(buffer, offset, plaintext, 0, count);
        
        var ciphertext = new byte[count];
        var tag = new byte[16];
        
        using var chacha = new ChaCha20Poly1305(_sessionKey);
        chacha.Encrypt(nonce, plaintext, ciphertext, tag);
        
        var packet = new byte[nonce.Length + ciphertext.Length + tag.Length];
        Buffer.BlockCopy(nonce, 0, packet, 0, nonce.Length);
        Buffer.BlockCopy(ciphertext, 0, packet, nonce.Length, ciphertext.Length);
        Buffer.BlockCopy(tag, 0, packet, nonce.Length + ciphertext.Length, tag.Length);
        
        var lengthBytes = BitConverter.GetBytes((uint)packet.Length);
        if (BitConverter.IsLittleEndian)
        {
            Array.Reverse(lengthBytes); // Convert to big-endian
        }
        
        await _innerStream.WriteAsync(lengthBytes, 0, 4, cancellationToken);
        await _innerStream.WriteAsync(packet, 0, packet.Length, cancellationToken);
    }

    public override int Read(byte[] buffer, int offset, int count)
    {
        return ReadAsync(buffer, offset, count).GetAwaiter().GetResult();
    }

    public override async Task<int> ReadAsync(byte[] buffer, int offset, int count, CancellationToken cancellationToken)
    {
        if (_recvBufferPos >= _recvBuffer.Length)
        {
            var lengthBytes = new byte[4];
            try
            {
                await ReadExactlyAsync(_innerStream, lengthBytes, cancellationToken);
            }
            catch (EndOfStreamException)
            {
                return 0;
            }
            
            if (BitConverter.IsLittleEndian)
            {
                Array.Reverse(lengthBytes);
            }
            var packetLength = BitConverter.ToUInt32(lengthBytes, 0);
            
            if (packetLength > 2 * 1024 * 1024) // 2MB max
            {
                throw new InvalidDataException($"Packet too large: {packetLength}");
            }
            
            var packet = new byte[packetLength];
            await ReadExactlyAsync(_innerStream, packet, cancellationToken);
            
            var nonce = new byte[12];
            Buffer.BlockCopy(packet, 0, nonce, 0, 12);
            
            var ciphertextLength = packet.Length - 12 - 16;
            var ciphertext = new byte[ciphertextLength];
            Buffer.BlockCopy(packet, 12, ciphertext, 0, ciphertextLength);
            
            var tag = new byte[16];
            Buffer.BlockCopy(packet, 12 + ciphertextLength, tag, 0, 16);
            
            var plaintext = new byte[ciphertextLength];
            using var chacha = new ChaCha20Poly1305(_sessionKey);
            chacha.Decrypt(nonce, ciphertext, tag, plaintext);
            
            _recvBuffer = plaintext;
            _recvBufferPos = 0;
        }
        
        var bytesToCopy = Math.Min(count, _recvBuffer.Length - _recvBufferPos);
        Buffer.BlockCopy(_recvBuffer, _recvBufferPos, buffer, offset, bytesToCopy);
        _recvBufferPos += bytesToCopy;
        
        return bytesToCopy;
    }

    private static async Task ReadExactlyAsync(Stream stream, byte[] buffer, CancellationToken cancellationToken)
    {
        int totalRead = 0;
        while (totalRead < buffer.Length)
        {
            int bytesRead = await stream.ReadAsync(buffer, totalRead, buffer.Length - totalRead, cancellationToken);
            if (bytesRead == 0)
            {
                throw new EndOfStreamException("Connection closed unexpectedly");
            }
            totalRead += bytesRead;
        }
    }

    public override void Flush() => _innerStream.Flush();
    public override Task FlushAsync(CancellationToken cancellationToken) => _innerStream.FlushAsync(cancellationToken);
    public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();
    public override void SetLength(long value) => throw new NotSupportedException();

    protected override void Dispose(bool disposing)
    {
        if (disposing)
        {
            _innerStream?.Dispose();
        }
        base.Dispose(disposing);
    }
}
`

func generateAuthHelper(logger *slog.Logger, projectDir string) error {
	logger.Debug("Generating ViiperAuth.cs")
	outputFile := filepath.Join(projectDir, "ViiperAuth.cs")

	if err := os.WriteFile(outputFile, []byte(authHelperTemplate), 0644); err != nil {
		return fmt.Errorf("write ViiperAuth.cs: %w", err)
	}

	logger.Info("Generated ViiperAuth.cs", "file", outputFile)
	return nil
}
