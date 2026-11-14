// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class Ns2proInput implements IBinarySerializable {
  Buttons!: number;
  Lx!: number;
  Ly!: number;
  Rx!: number;
  Ry!: number;
  Accelx!: number;
  Accely!: number;
  Accelz!: number;
  Gyrox!: number;
  Gyroy!: number;
  Gyroz!: number;

  constructor(init: Partial<Ns2proInput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeU32LE(this.Buttons as any);
    writer.writeU16LE(this.Lx as any);
    writer.writeU16LE(this.Ly as any);
    writer.writeU16LE(this.Rx as any);
    writer.writeU16LE(this.Ry as any);
    writer.writeI16LE(this.Accelx as any);
    writer.writeI16LE(this.Accely as any);
    writer.writeI16LE(this.Accelz as any);
    writer.writeI16LE(this.Gyrox as any);
    writer.writeI16LE(this.Gyroy as any);
    writer.writeI16LE(this.Gyroz as any);
  }
  static read(reader: BinaryReader): Ns2proInput {
    const buttons = reader.readU32LE();
    const lx = reader.readU16LE();
    const ly = reader.readU16LE();
    const rx = reader.readU16LE();
    const ry = reader.readU16LE();
    const accelx = reader.readI16LE();
    const accely = reader.readI16LE();
    const accelz = reader.readI16LE();
    const gyrox = reader.readI16LE();
    const gyroy = reader.readI16LE();
    const gyroz = reader.readI16LE();

		return new Ns2proInput({
      Buttons: buttons,
      Lx: lx,
      Ly: ly,
      Rx: rx,
      Ry: ry,
      Accelx: accelx,
      Accely: accely,
      Accelz: accelz,
      Gyrox: gyrox,
      Gyroy: gyroy,
      Gyroz: gyroz,
    });
	}
}
