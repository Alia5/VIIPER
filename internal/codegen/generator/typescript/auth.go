package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const authUtilsTemplate = `// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

import { Socket } from 'net';
import { createCipheriv, createDecipheriv, pbkdf2Sync, randomBytes, createHash, createHmac } from 'crypto';
import { Duplex } from 'stream';

const HANDSHAKE_MAGIC = 'eVI1\x00';
const NONCE_SIZE = 32;
const AUTH_CONTEXT = 'VIIPER-Auth-v1';
const SESSION_CONTEXT = 'VIIPER-Session-v1';
const PBKDF2_ITERATIONS = 100000;
const PBKDF2_SALT = 'VIIPER-Key-v1';

/**
 * Derive a 32-byte key from password using PBKDF2-SHA256
 */
function deriveKey(password: string): Buffer {
	if (!password || password.length === 0) {
		throw new Error('Password cannot be empty');
	}
	return pbkdf2Sync(password, PBKDF2_SALT, PBKDF2_ITERATIONS, 32, 'sha256');
}

/**
 * Derive session key from key and nonces using SHA-256
 */
function deriveSessionKey(key: Buffer, serverNonce: Buffer, clientNonce: Buffer): Buffer {
	const hash = createHash('sha256');
	hash.update(key);
	hash.update(serverNonce);
	hash.update(clientNonce);
	hash.update(Buffer.from(SESSION_CONTEXT));
	return hash.digest();
}

/**
 * Perform authentication handshake with VIIPER server
 * Returns a wrapped socket with encryption if handshake succeeds
 */
export async function performAuthHandshake(socket: Socket, password: string): Promise<Socket | EncryptedSocket> {
	const key = deriveKey(password);
	const clientNonce = randomBytes(NONCE_SIZE);
	const hmac = createHmac('sha256', key);
	hmac.update(Buffer.from(AUTH_CONTEXT));
	hmac.update(clientNonce);
	const authTag = hmac.digest();
	
	const handshakeMsg = Buffer.concat([
		Buffer.from(HANDSHAKE_MAGIC),
		clientNonce,
		authTag
	]);
	socket.write(handshakeMsg);
	
	const response = await readExactly(socket, 3 + NONCE_SIZE);
	
	const prefix = response.slice(0, 3).toString();
	if (prefix !== 'OK\x00') {
		const remaining = await readUntilEnd(socket);
		const fullResponse = Buffer.concat([response, remaining]).toString().trim();
		try {
			const error = JSON.parse(fullResponse);
			throw new Error(` + "`${error.status} ${error.title}: ${error.detail}`" + `);
		} catch {
			throw new Error(` + "`Invalid handshake response: ${fullResponse}`" + `);
		}
	}
	
	const serverNonce = response.slice(3);
	
	const sessionKey = deriveSessionKey(key, serverNonce, clientNonce);
	
	return new EncryptedSocket(socket, sessionKey);
}

/**
 * Read exact number of bytes from socket
 */
function readExactly(socket: Socket, length: number): Promise<Buffer> {
	return new Promise((resolve, reject) => {
		const chunks: Buffer[] = [];
		let received = 0;
		
		const onData = (chunk: Buffer) => {
			chunks.push(chunk);
			received += chunk.length;
			
			if (received >= length) {
				socket.removeListener('data', onData);
				socket.removeListener('error', onError);
				socket.removeListener('end', onEnd);
				
				const buffer = Buffer.concat(chunks);
				resolve(buffer.slice(0, length));
			}
		};
		
		const onError = (err: Error) => {
			socket.removeListener('data', onData);
			socket.removeListener('end', onEnd);
			reject(err);
		};
		
		const onEnd = () => {
			socket.removeListener('data', onData);
			socket.removeListener('error', onError);
			reject(new Error('Connection closed before receiving full response'));
		};
		
		socket.on('data', onData);
		socket.on('error', onError);
		socket.on('end', onEnd);
	});
}

/**
 * Read until socket closes
 */
function readUntilEnd(socket: Socket): Promise<Buffer> {
	return new Promise((resolve) => {
		const chunks: Buffer[] = [];
		socket.on('data', (chunk: Buffer) => chunks.push(chunk));
		socket.on('end', () => resolve(Buffer.concat(chunks)));
	});
}

/**
 * Encrypted socket wrapper using ChaCha20-Poly1305
 */
class EncryptedSocket extends Duplex {
	private socket: Socket;
	private sessionKey: Buffer;
	private sendCounter: number = 0;
	private recvBuffer: Buffer = Buffer.alloc(0);
	
	constructor(socket: Socket, sessionKey: Buffer) {
		super();
		this.socket = socket;
		this.sessionKey = sessionKey;
		
		socket.on('data', (chunk: Buffer) => this.handleIncomingData(chunk));
		socket.on('error', (err: Error) => this.emit('error', err));
		socket.on('end', () => this.emit('end'));
		socket.on('close', () => this.emit('close'));
	}
	
	_write(chunk: Buffer, encoding: string, callback: (error?: Error | null) => void): void {
		try {
			const nonce = Buffer.alloc(12);
			nonce.writeBigUInt64BE(BigInt(this.sendCounter), 4);
			this.sendCounter++;
			
			const cipher = createCipheriv('chacha20-poly1305', this.sessionKey, nonce, {
				authTagLength: 16
			});
			const ciphertext = Buffer.concat([cipher.update(chunk), cipher.final()]);
			const authTag = cipher.getAuthTag();

			const packet = Buffer.concat([nonce, ciphertext, authTag]);
			const lengthBuf = Buffer.alloc(4);
			lengthBuf.writeUInt32BE(packet.length, 0);
			
			this.socket.write(Buffer.concat([lengthBuf, packet]), callback);
		} catch (err) {
			callback(err as Error);
		}
	}
	
	_read(size: number): void {
		// Called when consumer wants to read, but we push data in handleIncomingData
	}
	
	private handleIncomingData(chunk: Buffer): void {
		this.recvBuffer = Buffer.concat([this.recvBuffer, chunk]);
		
		while (this.recvBuffer.length >= 4) {
			const packetLength = this.recvBuffer.readUInt32BE(0);
			
			if (this.recvBuffer.length < 4 + packetLength) {
				break;
			}
			
			const packet = this.recvBuffer.slice(4, 4 + packetLength);
			this.recvBuffer = this.recvBuffer.slice(4 + packetLength);
			
			try {
				const nonce = packet.slice(0, 12);
				const ciphertext = packet.slice(12, packet.length - 16);
				const authTag = packet.slice(packet.length - 16);
				
				const decipher = createDecipheriv('chacha20-poly1305', this.sessionKey, nonce, {
					authTagLength: 16
				});
				decipher.setAuthTag(authTag);
				
				const plaintext = Buffer.concat([decipher.update(ciphertext), decipher.final()]);
				
				this.push(plaintext);
			} catch (err) {
				this.emit('error', new Error(` + "`Decryption failed: ${(err as Error).message}`" + `));
			}
		}
	}
	
	setNoDelay(noDelay: boolean): this {
		this.socket.setNoDelay(noDelay);
		return this;
	}
	
	end(callback?: () => void): this {
		this.socket.end(callback);
		return this;
	}
	
	on(event: string | symbol, listener: (...args: any[]) => void): this {
		return super.on(event, listener);
	}
}

export { EncryptedSocket, deriveKey, deriveSessionKey };
`

func generateAuthUtils(logger *slog.Logger, utilsDir string) error {
	logger.Debug("Generating auth.ts")
	outputFile := filepath.Join(utilsDir, "auth.ts")

	if err := os.WriteFile(outputFile, []byte(authUtilsTemplate), 0644); err != nil {
		return fmt.Errorf("write auth.ts: %w", err)
	}

	logger.Info("Generated auth.ts", "file", outputFile)
	return nil
}
