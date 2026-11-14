// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class XboneOutput implements IBinarySerializable {
  Left!: number;
  Right!: number;
  Lefttrigger!: number;
  Righttrigger!: number;
  Duration!: number;
  Delay!: number;

  constructor(init: Partial<XboneOutput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU8(this.Left as any);
    writer.writeU8(this.Right as any);
    writer.writeU8(this.Lefttrigger as any);
    writer.writeU8(this.Righttrigger as any);
    writer.writeU8(this.Duration as any);
    writer.writeU8(this.Delay as any);
  }
  static read(reader: BinaryReader): XboneOutput {
    return new XboneOutput({
      Left: reader.readU8(),
      Right: reader.readU8(),
      Lefttrigger: reader.readU8(),
      Righttrigger: reader.readU8(),
      Duration: reader.readU8(),
      Delay: reader.readU8(),
    });
  }
}
