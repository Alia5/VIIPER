package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"
)

const binaryUtilsTemplate = `{{writeFileHeaderTS}}
export class BinaryWriter {
  private chunks: Buffer[] = [];

  writeU8(v: number): void { this.chunks.push(Buffer.from([v & 0xFF])); }
  writeI8(v: number): void { this.writeU8((v & 0xFF) >>> 0); }
  writeU16LE(v: number): void { const b = Buffer.allocUnsafe(2); b.writeUInt16LE(v); this.chunks.push(b); }
  writeI16LE(v: number): void { const b = Buffer.allocUnsafe(2); b.writeInt16LE(v); this.chunks.push(b); }
  writeU32LE(v: number): void { const b = Buffer.allocUnsafe(4); b.writeUInt32LE(v); this.chunks.push(b); }
  writeI32LE(v: number): void { const b = Buffer.allocUnsafe(4); b.writeInt32LE(v); this.chunks.push(b); }
  writeU64LE(v: bigint): void { const b = Buffer.allocUnsafe(8); b.writeBigUInt64LE(v); this.chunks.push(b); }
  writeI64LE(v: bigint): void { const b = Buffer.allocUnsafe(8); b.writeBigInt64LE(v); this.chunks.push(b); }
  writeBytes(buf: Buffer): void { this.chunks.push(buf); }

  toBuffer(): Buffer { return Buffer.concat(this.chunks); }
}

export class BinaryReader {
  private offset = 0;
  constructor(private buf: Buffer) {}
  readU8(): number { return this.buf.readUInt8(this.offset++); }
  readI8(): number { return this.buf.readInt8(this.offset++); }
  readU16LE(): number { const v = this.buf.readUInt16LE(this.offset); this.offset += 2; return v; }
  readI16LE(): number { const v = this.buf.readInt16LE(this.offset); this.offset += 2; return v; }
  readU32LE(): number { const v = this.buf.readUInt32LE(this.offset); this.offset += 4; return v; }
  readI32LE(): number { const v = this.buf.readInt32LE(this.offset); this.offset += 4; return v; }
  readU64LE(): bigint { const v = this.buf.readBigUInt64LE(this.offset); this.offset += 8; return v; }
  readI64LE(): bigint { const v = this.buf.readBigInt64LE(this.offset); this.offset += 8; return v; }
}
`

func generateBinaryUtils(logger *slog.Logger, utilsDir string) error {
	logger.Debug("Generating binary writer/reader utilities")
	f, err := os.Create(filepath.Join(utilsDir, "binary.ts"))
	if err != nil {
		return fmt.Errorf("write binary.ts: %w", err)
	}
	defer f.Close()
	tmpl := template.Must(template.New("binaryts").Funcs(template.FuncMap{
		"writeFileHeaderTS": writeFileHeaderTS,
	}).Parse(binaryUtilsTemplate))
	if err := tmpl.Execute(f, nil); err != nil {
		return fmt.Errorf("execute binary template: %w", err)
	}
	return nil
}
