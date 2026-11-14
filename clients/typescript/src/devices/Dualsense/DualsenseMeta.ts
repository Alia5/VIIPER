// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Typed deviceSpecific helpers for Dualsense


export interface MetaState {
  serialNumber: string;
  mACAddress: string;
  board: string;
  buildTime: string;
  batteryStatus: number;
  temperatureCelsius: number;
  batteryVoltage: number;
  shellColor: string;
  headphonesConnected: boolean;
  microphoneConnected: boolean;
}

export const metaStateToMap = (value: MetaState): Record<string, unknown> => ({
  serial_number: value.serialNumber,
  mac_address: value.mACAddress,
  board: value.board,
  build_time: value.buildTime,
  battery_status: value.batteryStatus,
  temperature_celsius: value.temperatureCelsius,
  battery_voltage: value.batteryVoltage,
  shell_color: value.shellColor,
  headphones_connected: value.headphonesConnected,
  microphone_connected: value.microphoneConnected,
});

export const metaStateFromMap = (data: Record<string, unknown> | undefined | null): MetaState => {
  const map = data ?? {};
  return {
    serialNumber: map['serial_number'] as string,
    mACAddress: map['mac_address'] as string,
    board: map['board'] as string,
    buildTime: map['build_time'] as string,
    batteryStatus: map['battery_status'] as number,
    temperatureCelsius: map['temperature_celsius'] as number,
    batteryVoltage: map['battery_voltage'] as number,
    shellColor: map['shell_color'] as string,
    headphonesConnected: map['headphones_connected'] as boolean,
    microphoneConnected: map['microphone_connected'] as boolean,
  };
};


