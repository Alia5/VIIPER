// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class Ns2proOutput implements IBinarySerializable {
  Leftrumble!: number[];
  Rightrumble!: number[];
  Flags!: number;
  Playerledmask!: number;

  constructor(init: Partial<Ns2proOutput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {

		{
			const arr = (this.Leftrumble ?? []) as any[];
			for (let i = 0; i < 16; i++) {
				writer.writeU8((arr[i] ?? 0) as any);
			}
		}

		{
			const arr = (this.Rightrumble ?? []) as any[];
			for (let i = 0; i < 16; i++) {
				writer.writeU8((arr[i] ?? 0) as any);
			}
		}
    writer.writeU8(this.Flags as any);
    writer.writeU8(this.Playerledmask as any);
  }
  static read(reader: BinaryReader): Ns2proOutput {
    const leftrumble: any[] = new Array(16);
		for (let i = 0; i < 16; i++) {
			leftrumble[i] = reader.readU8();
		}
    const rightrumble: any[] = new Array(16);
		for (let i = 0; i < 16; i++) {
			rightrumble[i] = reader.readU8();
		}
    const flags = reader.readU8();
    const playerledmask = reader.readU8();

		return new Ns2proOutput({
      Leftrumble: leftrumble,
      Rightrumble: rightrumble,
      Flags: flags,
      Playerledmask: playerledmask,
    });
	}
}
