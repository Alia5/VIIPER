import {
  TriggerMode,
  TriggerZoneCount,
} from './DualsenseFullConstants';

const bitsPerStrength = 3;
const strengthMask = 0x07;
const frequencyParam = 8;

export interface TriggerEffect {
  mode: number;
  params: number[];
  zoneStrengths: number[];
  startZone: number;
  endZone: number;
  startPosition: number;
  endPosition: number;
  strength: number;
  snapForce: number;
  amplitudeA: number;
  amplitudeB: number;
  firstFoot: number;
  secondFoot: number;
  frequency: number;
  period: number;
}

export const decodeTriggerEffect = (mode: number, params: number[]): TriggerEffect => {
  const effect: TriggerEffect = {
    mode,
    params,
    zoneStrengths: new Array<number>(TriggerZoneCount).fill(0),
    startZone: 0,
    endZone: 0,
    startPosition: 0,
    endPosition: 0,
    strength: 0,
    snapForce: 0,
    amplitudeA: 0,
    amplitudeB: 0,
    firstFoot: 0,
    secondFoot: 0,
    frequency: 0,
    period: 0,
  };
  const zones = params[0] | (params[1] << 8);

  switch (mode) {
    case TriggerMode.Feedback:
    case TriggerMode.Vibration: {
      const packedStrengths = (params[2] | (params[3] << 8) | (params[4] << 16) | (params[5] << 24)) >>> 0;
      for (let zone = 0; zone < TriggerZoneCount; zone++) {
        if (zones & (1 << zone)) {
          effect.zoneStrengths[zone] = ((packedStrengths >>> (bitsPerStrength * zone)) & strengthMask) + 1;
        }
      }
      if (mode === TriggerMode.Vibration) {
        effect.frequency = params[frequencyParam];
      }
      break;
    }
    case TriggerMode.Weapon:
    case TriggerMode.Bow:
    case TriggerMode.Galloping:
    case TriggerMode.Machine: {
      for (let zone = TriggerZoneCount - 1; zone >= 0; zone--) {
        if (zones & (1 << zone)) {
          effect.startZone = zone;
        }
      }
      for (let zone = 0; zone < TriggerZoneCount; zone++) {
        if (zones & (1 << zone)) {
          effect.endZone = zone;
        }
      }
      const first = params[2] & strengthMask;
      const second = (params[2] >> bitsPerStrength) & strengthMask;
      if (mode === TriggerMode.Weapon) {
        effect.strength = first + 1;
      }
      if (mode === TriggerMode.Bow) {
        effect.strength = first + 1;
        effect.snapForce = second + 1;
      }
      if (mode === TriggerMode.Galloping) {
        effect.secondFoot = first;
        effect.firstFoot = second;
        effect.frequency = params[3];
      }
      if (mode === TriggerMode.Machine) {
        effect.amplitudeA = first;
        effect.amplitudeB = second;
        effect.frequency = params[3];
        effect.period = params[4];
      }
      break;
    }
    case TriggerMode.SimpleFeedback:
    case TriggerMode.LimitedFeedback:
      effect.startPosition = params[0];
      effect.strength = params[1];
      break;
    case TriggerMode.SimpleWeapon:
    case TriggerMode.LimitedWeapon:
      effect.startPosition = params[0];
      effect.endPosition = params[1];
      effect.strength = params[2];
      break;
    case TriggerMode.SimpleVibration:
      effect.frequency = params[0];
      effect.strength = params[1];
      effect.startPosition = params[2];
      break;
  }
  return effect;
};
