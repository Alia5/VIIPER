// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class KeyboardInput implements IBinarySerializable {
  Modifiers!: number;
  Count!: number;
  Keys!: number[];

  constructor(init: Partial<KeyboardInput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Modifiers as any);
    writer.writeU8(this.Count as any);

		for (let i = 0; i < Number((this.Count as any)); i++) {
			writer.writeU8((this.Keys as any[])[i]);
		}
  }
  static read(reader: BinaryReader): KeyboardInput {
    const modifiers = reader.readU8();
    const count = reader.readU8();
    const keys: any[] = new Array(Number(count));
		for (let i = 0; i < Number(count); i++) {
			keys[i] = reader.readU8();
		}

		return new KeyboardInput({
      Modifiers: modifiers,
      Count: count,
      Keys: keys,
    });
	}
}
