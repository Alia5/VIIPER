// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class KeyboardOutput implements IBinarySerializable {
  Leds!: number;

  constructor(init: Partial<KeyboardOutput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Leds as any);
  }
  static read(reader: BinaryReader): KeyboardOutput {
    const leds = reader.readU8();

		return new KeyboardOutput({
      Leds: leds,
    });
	}
}
