import {
  MetaState,
  metaStateToMap,
} from '../Dualsense/DualsenseMeta';
import { OutputStateSize } from '../Dualsense/DualsenseConstants';

export const DeviceType = 'dualsense_full';
export const DeviceTypeEdge = 'dualsenseedge_full';

export const FullOutputStateSize = OutputStateSize + 6;
export const PowerSaveControlMicrophoneMute = 0x10;
export const AudioControlOutputPath = 0x30;

export const OutputReportOffsetSpeakerVolume = 6;
export const OutputReportOffsetMicrophoneVolume = 7;
export const OutputReportOffsetAudioControl = 8;
export const OutputReportOffsetMicMuteLed = 9;
export const OutputReportOffsetPowerSaveControl = 10;
export const OutputReportOffsetAudioControl2 = 38;

export const AudioSampleRate = 48000;
export const AudioBytesPerSample = 2;
export const AudioOutputChannels = 4;
export const AudioMicrophoneChannels = 2;
export const AudioOutputPacketSize = AudioSampleRate / 1000 * AudioOutputChannels * AudioBytesPerSample;
export const AudioMicrophonePacketSize = AudioSampleRate / 1000 * AudioMicrophoneChannels * AudioBytesPerSample;

export interface CreateOptions {
  meta?: MetaState;
  audioPort?: number;
}

export const createOptionsToMap = (value: CreateOptions): Record<string, unknown> => {
  const map: Record<string, unknown> = {};
  if (value.meta !== undefined) {
    map['Meta'] = metaStateToMap(value.meta);
  }
  if (value.audioPort !== undefined) {
    map['audio_port'] = value.audioPort;
  }
  return map;
};
