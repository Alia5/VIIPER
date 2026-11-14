// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';

export class DualsenseInput implements IBinarySerializable {
  Sticklx!: number;
  Stickly!: number;
  Stickrx!: number;
  Stickry!: number;
  Buttons!: number;
  Dpad!: number;
  Triggerl2!: number;
  Triggerr2!: number;
  Touch1x!: number;
  Touch1y!: number;
  Touch1active!: number;
  Touch2x!: number;
  Touch2y!: number;
  Touch2active!: number;
  Gyrox!: number;
  Gyroy!: number;
  Gyroz!: number;
  Accelx!: number;
  Accely!: number;
  Accelz!: number;

  constructor(init: Partial<DualsenseInput> = {}) {
    Object.assign(this, init);
  }
  write(writer: BinaryWriter): void {
    writer.writeI8(this.Sticklx as any);
    writer.writeI8(this.Stickly as any);
    writer.writeI8(this.Stickrx as any);
    writer.writeI8(this.Stickry as any);
    writer.writeU32LE(this.Buttons as any);
    writer.writeU8(this.Dpad as any);
    writer.writeU8(this.Triggerl2 as any);
    writer.writeU8(this.Triggerr2 as any);
    writer.writeU16LE(this.Touch1x as any);
    writer.writeU16LE(this.Touch1y as any);
    writer.writeU8(this.Touch1active as any);
    writer.writeU16LE(this.Touch2x as any);
    writer.writeU16LE(this.Touch2y as any);
    writer.writeU8(this.Touch2active as any);
    writer.writeI16LE(this.Gyrox as any);
    writer.writeI16LE(this.Gyroy as any);
    writer.writeI16LE(this.Gyroz as any);
    writer.writeI16LE(this.Accelx as any);
    writer.writeI16LE(this.Accely as any);
    writer.writeI16LE(this.Accelz as any);
  }
  static read(reader: BinaryReader): DualsenseInput {
    const sticklx = reader.readI8();
    const stickly = reader.readI8();
    const stickrx = reader.readI8();
    const stickry = reader.readI8();
    const buttons = reader.readU32LE();
    const dpad = reader.readU8();
    const triggerl2 = reader.readU8();
    const triggerr2 = reader.readU8();
    const touch1x = reader.readU16LE();
    const touch1y = reader.readU16LE();
    const touch1active = reader.readU8();
    const touch2x = reader.readU16LE();
    const touch2y = reader.readU16LE();
    const touch2active = reader.readU8();
    const gyrox = reader.readI16LE();
    const gyroy = reader.readI16LE();
    const gyroz = reader.readI16LE();
    const accelx = reader.readI16LE();
    const accely = reader.readI16LE();
    const accelz = reader.readI16LE();

		return new DualsenseInput({
      Sticklx: sticklx,
      Stickly: stickly,
      Stickrx: stickrx,
      Stickry: stickry,
      Buttons: buttons,
      Dpad: dpad,
      Triggerl2: triggerl2,
      Triggerr2: triggerr2,
      Touch1x: touch1x,
      Touch1y: touch1y,
      Touch1active: touch1active,
      Touch2x: touch2x,
      Touch2y: touch2y,
      Touch2active: touch2active,
      Gyrox: gyrox,
      Gyroy: gyroy,
      Gyroz: gyroz,
      Accelx: accelx,
      Accely: accely,
      Accelz: accelz,
    });
	}
}
