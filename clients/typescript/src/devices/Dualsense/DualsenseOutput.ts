// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class DualsenseOutput implements IBinarySerializable {
  Updateflags1!: number;
  Updateflags2!: number;
  Updateflags3!: number;
  Rumblesmall!: number;
  Rumblelarge!: number;
  Ledred!: number;
  Ledgreen!: number;
  Ledblue!: number;
  Playerleds!: number;

  constructor(init: Partial<DualsenseOutput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Updateflags1 as any);
    writer.writeU8(this.Updateflags2 as any);
    writer.writeU8(this.Updateflags3 as any);
    writer.writeU8(this.Rumblesmall as any);
    writer.writeU8(this.Rumblelarge as any);
    writer.writeU8(this.Ledred as any);
    writer.writeU8(this.Ledgreen as any);
    writer.writeU8(this.Ledblue as any);
    writer.writeU8(this.Playerleds as any);
  }
  static read(reader: BinaryReader): DualsenseOutput {
    const updateflags1 = reader.readU8();
    const updateflags2 = reader.readU8();
    const updateflags3 = reader.readU8();
    const rumblesmall = reader.readU8();
    const rumblelarge = reader.readU8();
    const ledred = reader.readU8();
    const ledgreen = reader.readU8();
    const ledblue = reader.readU8();
    const playerleds = reader.readU8();

		return new DualsenseOutput({
      Updateflags1: updateflags1,
      Updateflags2: updateflags2,
      Updateflags3: updateflags3,
      Rumblesmall: rumblesmall,
      Rumblelarge: rumblelarge,
      Ledred: ledred,
      Ledgreen: ledgreen,
      Ledblue: ledblue,
      Playerleds: playerleds,
    });
	}
}
