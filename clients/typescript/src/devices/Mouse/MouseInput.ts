// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class MouseInput implements IBinarySerializable {
  Buttons!: number;
  Dx!: number;
  Dy!: number;
  Wheel!: number;
  Pan!: number;

  constructor(init: Partial<MouseInput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Buttons as any);
    writer.writeI16LE(this.Dx as any);
    writer.writeI16LE(this.Dy as any);
    writer.writeI16LE(this.Wheel as any);
    writer.writeI16LE(this.Pan as any);
  }
  static read(reader: BinaryReader): MouseInput {
    const buttons = reader.readU8();
    const dx = reader.readI16LE();
    const dy = reader.readI16LE();
    const wheel = reader.readI16LE();
    const pan = reader.readI16LE();

		return new MouseInput({
      Buttons: buttons,
      Dx: dx,
      Dy: dy,
      Wheel: wheel,
      Pan: pan,
    });
	}
}
