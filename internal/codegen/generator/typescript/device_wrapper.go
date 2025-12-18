package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func generateDeviceWrapper(logger *slog.Logger, srcDir string) error {
	logger.Debug("Generating ViiperDevice.ts stream wrapper")
	content := writeFileHeaderTS() + `import { Socket } from 'net';
import { BinaryWriter } from './utils/binary';
import { EventEmitter } from 'events';

export interface IBinarySerializable {
  write(writer: BinaryWriter): void;
}

export class ViiperDevice extends EventEmitter {
  private socket: Socket;

  constructor(socket: Socket) {
    super();
    this.socket = socket;
    this.socket.on('data', (buf: Buffer) => {
      // emit raw output frames; higher-level parsers can decode as needed
      this.emit('output', buf);
    });
    this.socket.on('error', (err: Error) => {
      this.emit('error', err);
    });
    this.socket.on('end', () => {
      this.emit('end');
    });
  }

  async send<T extends IBinarySerializable>(payload: T): Promise<void> {
    const writer = new BinaryWriter();
    payload.write(writer);
    const buf = writer.toBuffer();
    await new Promise<void>((resolve, reject) => {
      this.socket.write(buf, (err) => err ? reject(err as Error) : resolve());
    });
  }

  async sendRaw(buf: Buffer): Promise<void> {
    await new Promise<void>((resolve, reject) => {
      this.socket.write(buf, (err) => err ? reject(err as Error) : resolve());
    });
  }

  close(): void {
    this.socket.end();
    this.removeAllListeners();
  }
}
`
	if err := os.WriteFile(filepath.Join(srcDir, "ViiperDevice.ts"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write ViiperDevice.ts: %w", err)
	}
	return nil
}
