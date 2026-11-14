// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class Xbox360Input implements IBinarySerializable {
  Buttons!: number;
  Lt!: number;
  Rt!: number;
  Lx!: number;
  Ly!: number;
  Rx!: number;
  Ry!: number;
  Reserved!: number[];

  constructor(init: Partial<Xbox360Input> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU32LE(this.Buttons as any);
    writer.writeU8(this.Lt as any);
    writer.writeU8(this.Rt as any);
    writer.writeI16LE(this.Lx as any);
    writer.writeI16LE(this.Ly as any);
    writer.writeI16LE(this.Rx as any);
    writer.writeI16LE(this.Ry as any);

		{
			const arr = (this.Reserved ?? []) as any[];
			for (let i = 0; i < 6; i++) {
				writer.writeU8((arr[i] ?? 0) as any);
			}
		}
  }
  static read(reader: BinaryReader): Xbox360Input {
    const buttons = reader.readU32LE();
    const lt = reader.readU8();
    const rt = reader.readU8();
    const lx = reader.readI16LE();
    const ly = reader.readI16LE();
    const rx = reader.readI16LE();
    const ry = reader.readI16LE();
    const reserved: any[] = new Array(6);
		for (let i = 0; i < 6; i++) {
			reserved[i] = reader.readU8();
		}

		return new Xbox360Input({
      Buttons: buttons,
      Lt: lt,
      Rt: rt,
      Lx: lx,
      Ly: ly,
      Rx: rx,
      Ry: ry,
      Reserved: reserved,
    });
	}
}
