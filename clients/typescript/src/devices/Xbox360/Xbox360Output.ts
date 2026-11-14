// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class Xbox360Output implements IBinarySerializable {
  Left!: number;
  Right!: number;

  constructor(init: Partial<Xbox360Output> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Left as any);
    writer.writeU8(this.Right as any);
  }
  static read(reader: BinaryReader): Xbox360Output {
    const left = reader.readU8();
    const right = reader.readU8();

		return new Xbox360Output({
      Left: left,
      Right: right,
    });
	}
}
