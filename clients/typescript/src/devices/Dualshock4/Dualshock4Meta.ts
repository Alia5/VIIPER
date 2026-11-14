// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Typed deviceSpecific helpers for Dualshock4


export interface MetaState {
  serialNumber: string;
  board: string;
  buildTime: string;
  batteryStatus: number;
  temperatureCelsius: number;
  batteryVoltage: number;
}

export const metaStateToMap = (value: MetaState): Record<string, unknown> => ({
  serial_number: value.serialNumber,
  board: value.board,
  build_time: value.buildTime,
  battery_status: value.batteryStatus,
  temperature_celsius: value.temperatureCelsius,
  battery_voltage: value.batteryVoltage,
});

export const metaStateFromMap = (data: Record<string, unknown> | undefined | null): MetaState => {
  const map = data ?? {};
  return {
    serialNumber: map['serial_number'] as string,
    board: map['board'] as string,
    buildTime: map['build_time'] as string,
    batteryStatus: map['battery_status'] as number,
    temperatureCelsius: map['temperature_celsius'] as number,
    batteryVoltage: map['battery_voltage'] as number,
  };
};


