// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class Dualshock4Output implements IBinarySerializable {
  Updateflags!: number;
  Rumblesmall!: number;
  Rumblelarge!: number;
  Ledred!: number;
  Ledgreen!: number;
  Ledblue!: number;
  Flashon!: number;
  Flashoff!: number;

  constructor(init: Partial<Dualshock4Output> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Updateflags as any);
    writer.writeU8(this.Rumblesmall as any);
    writer.writeU8(this.Rumblelarge as any);
    writer.writeU8(this.Ledred as any);
    writer.writeU8(this.Ledgreen as any);
    writer.writeU8(this.Ledblue as any);
    writer.writeU8(this.Flashon as any);
    writer.writeU8(this.Flashoff as any);
  }
  static read(reader: BinaryReader): Dualshock4Output {
    const updateflags = reader.readU8();
    const rumblesmall = reader.readU8();
    const rumblelarge = reader.readU8();
    const ledred = reader.readU8();
    const ledgreen = reader.readU8();
    const ledblue = reader.readU8();
    const flashon = reader.readU8();
    const flashoff = reader.readU8();

		return new Dualshock4Output({
      Updateflags: updateflags,
      Rumblesmall: rumblesmall,
      Rumblelarge: rumblelarge,
      Ledred: ledred,
      Ledgreen: ledgreen,
      Ledblue: ledblue,
      Flashon: flashon,
      Flashoff: flashoff,
    });
	}
}
