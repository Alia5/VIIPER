// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class XboneInput implements IBinarySerializable {
  Buttons!: number;
  Lt!: number;
  Rt!: number;
  Lx!: number;
  Ly!: number;
  Rx!: number;
  Ry!: number;

  constructor(init: Partial<XboneInput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU32LE(this.Buttons as any);
    writer.writeU16LE(this.Lt as any);
    writer.writeU16LE(this.Rt as any);
    writer.writeI16LE(this.Lx as any);
    writer.writeI16LE(this.Ly as any);
    writer.writeI16LE(this.Rx as any);
    writer.writeI16LE(this.Ry as any);
  }
  static read(reader: BinaryReader): XboneInput {
    return new XboneInput({
      Buttons: reader.readU32LE(),
      Lt: reader.readU16LE(),
      Rt: reader.readU16LE(),
      Lx: reader.readI16LE(),
      Ly: reader.readI16LE(),
      Rx: reader.readI16LE(),
      Ry: reader.readI16LE(),
    });
  }
}
