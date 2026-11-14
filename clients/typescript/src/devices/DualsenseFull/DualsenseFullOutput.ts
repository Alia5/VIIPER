import {
  BinaryWriter,
  BinaryReader,
} from '../../utils/binary';
import { DualsenseOutput } from '../Dualsense/DualsenseOutput';
import {
  OutputReportOffsetAudioControl,
  OutputReportOffsetAudioControl2,
  OutputReportOffsetMicMuteLed,
  OutputReportOffsetMicrophoneVolume,
  OutputReportOffsetPowerSaveControl,
  OutputReportOffsetSpeakerVolume,
} from './DualsenseFullConstants';

export class DualsenseFullOutput extends DualsenseOutput {
  Speakervolume!: number;
  Microphonevolume!: number;
  Audiocontrol!: number;
  Micmuteled!: number;
  Powersavecontrol!: number;
  Audiocontrol2!: number;

  constructor(init: Partial<DualsenseFullOutput> = {}) {
    super(init);
    Object.assign(this, init);
  }

  outputReport(): Uint8Array {
    const report = super.outputReport();
    report[OutputReportOffsetSpeakerVolume] = this.Speakervolume;
    report[OutputReportOffsetMicrophoneVolume] = this.Microphonevolume;
    report[OutputReportOffsetAudioControl] = this.Audiocontrol;
    report[OutputReportOffsetMicMuteLed] = this.Micmuteled;
    report[OutputReportOffsetPowerSaveControl] = this.Powersavecontrol;
    report[OutputReportOffsetAudioControl2] = this.Audiocontrol2;
    return report;
  }

  write(writer: BinaryWriter): void {
    super.write(writer);
    writer.writeU8(this.Speakervolume);
    writer.writeU8(this.Microphonevolume);
    writer.writeU8(this.Audiocontrol);
    writer.writeU8(this.Micmuteled);
    writer.writeU8(this.Powersavecontrol);
    writer.writeU8(this.Audiocontrol2);
  }

  static read(reader: BinaryReader): DualsenseFullOutput {
    const base = DualsenseOutput.read(reader);
    return new DualsenseFullOutput({
      ...base,
      Speakervolume: reader.readU8(),
      Microphonevolume: reader.readU8(),
      Audiocontrol: reader.readU8(),
      Micmuteled: reader.readU8(),
      Powersavecontrol: reader.readU8(),
      Audiocontrol2: reader.readU8(),
    });
  }
}
