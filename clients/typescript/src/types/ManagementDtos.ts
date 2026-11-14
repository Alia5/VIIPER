// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


// Management API DTO interfaces


// APIError DTO
export interface APIError {
  status: number;

  title: string;

  detail: string;

}


// PingResponse DTO
export interface PingResponse {
  server: string;

  version: string;

}


// BusListResponse DTO
export interface BusListResponse {
  buses: number[];

}


// BusCreateResponse DTO
export interface BusCreateResponse {
  busId: number;

}


// BusRemoveResponse DTO
export interface BusRemoveResponse {
  busId: number;

}


// Device DTO
export interface Device {
  busId: number;

  devId: string;

  vid: string;

  pid: string;

  type: string;

  deviceSpecific: Record<string, unknown>;

}


// DevicesListResponse DTO
export interface DevicesListResponse {
  devices: Device[];

}


// DeviceRemoveResponse DTO
export interface DeviceRemoveResponse {
  busId: number;

  devId: string;

}


// DeviceCreateRequest DTO
export interface DeviceCreateRequest {
  type?: string;

  idVendor?: number;

  idProduct?: number;

  deviceSpecific?: Record<string, unknown>;

}


