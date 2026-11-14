// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Typed deviceSpecific helpers for Ns2pro


export interface MetaState {
  serialNumber: string;
  batteryLevel: number;
  charging: boolean;
  externalPower: boolean;
  batteryVolts: number;
}

export const metaStateToMap = (value: MetaState): Record<string, unknown> => ({
  serial_number: value.serialNumber,
  battery_level: value.batteryLevel,
  charging: value.charging,
  external_power: value.externalPower,
  battery_volts: value.batteryVolts,
});

export const metaStateFromMap = (data: Record<string, unknown> | undefined | null): MetaState => {
  const map = data ?? {};
  return {
    serialNumber: map['serial_number'] as string,
    batteryLevel: map['battery_level'] as number,
    charging: map['charging'] as boolean,
    externalPower: map['external_power'] as boolean,
    batteryVolts: map['battery_volts'] as number,
  };
};


