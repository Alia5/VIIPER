import {
  BinaryWriter,
  BinaryReader,
} from '../../utils/binary';
import { DualsenseOutput } from '../Dualsense/DualsenseOutput';
import { TriggerParamsSize } from './DualsenseFullConstants';
import {
  TriggerEffect,
  decodeTriggerEffect,
} from './DualsenseFullTriggerEffect';

export class DualsenseFullOutput extends DualsenseOutput {
  TriggerR2Mode!: number;
  TriggerR2Params!: number[];
  TriggerL2Mode!: number;
  TriggerL2Params!: number[];

  constructor(init: Partial<DualsenseFullOutput> = {}) {
    super(init);
    Object.assign(this, init);
  }

  triggerR2(): TriggerEffect {
    return decodeTriggerEffect(this.TriggerR2Mode, this.TriggerR2Params);
  }

  triggerL2(): TriggerEffect {
    return decodeTriggerEffect(this.TriggerL2Mode, this.TriggerL2Params);
  }

  write(writer: BinaryWriter): void {
    super.write(writer);
    writer.writeU8(this.TriggerR2Mode);
    for (let i = 0; i < TriggerParamsSize; i++) {
      writer.writeU8(this.TriggerR2Params?.[i] ?? 0);
    }
    writer.writeU8(this.TriggerL2Mode);
    for (let i = 0; i < TriggerParamsSize; i++) {
      writer.writeU8(this.TriggerL2Params?.[i] ?? 0);
    }
  }

  static read(reader: BinaryReader): DualsenseFullOutput {
    const base = DualsenseOutput.read(reader);
    const triggerR2Mode = reader.readU8();
    const triggerR2Params: number[] = [];
    for (let i = 0; i < TriggerParamsSize; i++) {
      triggerR2Params.push(reader.readU8());
    }
    const triggerL2Mode = reader.readU8();
    const triggerL2Params: number[] = [];
    for (let i = 0; i < TriggerParamsSize; i++) {
      triggerL2Params.push(reader.readU8());
    }

    return new DualsenseFullOutput({
      ...base,
      TriggerR2Mode: triggerR2Mode,
      TriggerR2Params: triggerR2Params,
      TriggerL2Mode: triggerL2Mode,
      TriggerL2Params: triggerL2Params,
    });
  }
}
