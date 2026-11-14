import { BinaryWriter, BinaryReader } from '../../utils/binary';
import type { IBinarySerializable } from '../../ViiperDevice';
import { TriggerParamsSize } from './DualsenseConstants';
import {
  TriggerEffect,
  decodeTriggerEffect,
} from './DualsenseTriggerEffect';

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
  TriggerR2Mode!: number;
  TriggerR2Params!: number[];
  TriggerL2Mode!: number;
  TriggerL2Params!: number[];

  constructor(init: Partial<DualsenseOutput> = {}) {
    Object.assign(this, init);
  }

  triggerR2(): TriggerEffect {
    return decodeTriggerEffect(this.TriggerR2Mode, this.TriggerR2Params);
  }

  triggerL2(): TriggerEffect {
    return decodeTriggerEffect(this.TriggerL2Mode, this.TriggerL2Params);
  }

  write(writer: BinaryWriter): void {
    writer.writeU8(this.Updateflags1);
    writer.writeU8(this.Updateflags2);
    writer.writeU8(this.Updateflags3);
    writer.writeU8(this.Rumblesmall);
    writer.writeU8(this.Rumblelarge);
    writer.writeU8(this.Ledred);
    writer.writeU8(this.Ledgreen);
    writer.writeU8(this.Ledblue);
    writer.writeU8(this.Playerleds);
    writer.writeU8(this.TriggerR2Mode);
    for (let i = 0; i < TriggerParamsSize; i++) {
      writer.writeU8(this.TriggerR2Params?.[i] ?? 0);
    }
    writer.writeU8(this.TriggerL2Mode);
    for (let i = 0; i < TriggerParamsSize; i++) {
      writer.writeU8(this.TriggerL2Params?.[i] ?? 0);
    }
  }

  static read(reader: BinaryReader): DualsenseOutput {
    const output = new DualsenseOutput({
      Updateflags1: reader.readU8(),
      Updateflags2: reader.readU8(),
      Updateflags3: reader.readU8(),
      Rumblesmall: reader.readU8(),
      Rumblelarge: reader.readU8(),
      Ledred: reader.readU8(),
      Ledgreen: reader.readU8(),
      Ledblue: reader.readU8(),
      Playerleds: reader.readU8(),
      TriggerR2Params: [],
      TriggerL2Params: [],
    });
    output.TriggerR2Mode = reader.readU8();
    for (let i = 0; i < TriggerParamsSize; i++) {
      output.TriggerR2Params.push(reader.readU8());
    }
    output.TriggerL2Mode = reader.readU8();
    for (let i = 0; i < TriggerParamsSize; i++) {
      output.TriggerL2Params.push(reader.readU8());
    }
    return output;
  }
}
