// Auto-generated VIIPER TypeScript Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase


import { Socket } from 'net';
import { TextDecoder, TextEncoder } from 'util';
import type * as Types from './types/ManagementDtos';
import { ViiperDevice } from './ViiperDevice';
import { performAuthHandshake } from './utils/auth';

const encoder = new TextEncoder();
const decoder = new TextDecoder();

/**
 * VIIPER management & streaming API client.
 * Request framing: <path>[ <payload>]\0 (null terminator) ; Response framing: single JSON line ending in \n then connection close.
 * 
 * @param host - VIIPER server hostname or IP address
 * @param port - VIIPER API server port (default: 3242)
 * @param password - Authentication password (default: "" = no auth). Empty string explicitly means no authentication.
 */
export class ViiperClient {
	private host: string;
	private port: number;
	private password: string;

	constructor(host: string, port: number = 3242, password: string = "") {
		this.host = host;
		this.port = port;
		this.password = password;
	}

	/**
	 * Ping: ping
	 */
	async ping(): Promise<Types.PingResponse> {
		const path = `ping`;
		const payload: string = '';
		return await this.sendRequest<Types.PingResponse>(path, payload);
	}

	/**
	 * BusList: bus/list
	 */
	async buslist(): Promise<Types.BusListResponse> {
		const path = `bus/list`;
		const payload: string = '';
		return await this.sendRequest<Types.BusListResponse>(path, payload);
	}

	/**
	 * BusCreate: bus/create
	 */
	async buscreate(id?: number): Promise<Types.BusCreateResponse> {
		const path = `bus/create`;
		const payload: string = id !== undefined && id !== null ? String(id) : '';
		return await this.sendRequest<Types.BusCreateResponse>(path, payload);
	}

	/**
	 * BusRemove: bus/remove
	 */
	async busremove(id: number): Promise<Types.BusRemoveResponse> {
		const path = `bus/remove`;
		const payload: string = id !== undefined && id !== null ? String(id) : '';
		return await this.sendRequest<Types.BusRemoveResponse>(path, payload);
	}

	/**
	 * BusDevicesList: bus/{id}/list
	 */
	async busdeviceslist(id: number): Promise<Types.DevicesListResponse> {
		const path = `bus/{id}/list`.replace("{id}", String(id));
		const payload: string = '';
		return await this.sendRequest<Types.DevicesListResponse>(path, payload);
	}

	/**
	 * BusDeviceAdd: bus/{id}/add
	 */
	async busdeviceadd(id: number, devicecreaterequest: Types.DeviceCreateRequest): Promise<Types.Device> {
		const path = `bus/{id}/add`.replace("{id}", String(id));
		const payload: string = JSON.stringify(devicecreaterequest);
		return await this.sendRequest<Types.Device>(path, payload);
	}

	/**
	 * BusDeviceRemove: bus/{id}/remove
	 */
	async busdeviceremove(id: number, value: string): Promise<Types.DeviceRemoveResponse> {
		const path = `bus/{id}/remove`.replace("{id}", String(id));
		const payload: string = value ? String(value) : '';
		return await this.sendRequest<Types.DeviceRemoveResponse>(path, payload);
	}

	private async sendRequest<T>(path: string, payload?: string | null): Promise<T> {
		return new Promise<T>(async (resolve, reject) => {
			const socket = new Socket();
			socket.connect(this.port, this.host, async () => {
				try {
					socket.setNoDelay(true);
					
					let wrappedSocket: Socket | any = socket;
					if (this.password) {
						wrappedSocket = await performAuthHandshake(socket, this.password);
					}
					
					let line = path; // preserve case
					if (payload && payload.length > 0) line += ' ' + payload;
					line += '\0';
					wrappedSocket.write(encoder.encode(line));
					
					let buffer = '';
					const handleData = (chunk: Buffer) => {
						buffer += decoder.decode(chunk);
						const nlIdx = buffer.indexOf('\n');
						if (nlIdx !== -1) {
							const jsonLine = buffer.slice(0, nlIdx);
							let parsed: any;
							try {
								parsed = JSON.parse(jsonLine);
							} catch (e) {
								wrappedSocket.end();
								reject(e);
								return;
							}
							if (parsed && typeof parsed === 'object' && 'status' in parsed && parsed.status >= 400) {
								wrappedSocket.end();
								reject(new Error(String(parsed.status) + ' ' + parsed.title + ': ' + parsed.detail));
								return;
							}
							wrappedSocket.end();
							resolve(parsed as T);
						}
					};
					
					wrappedSocket.on('data', handleData);
				} catch (e) {
					socket.end();
					reject(e);
				}
			});

			socket.on('error', reject);
			socket.on('end', () => {/* noop */});
		});
	}

	async connectDevice(busId: number, devId: string): Promise<ViiperDevice> {
		return new Promise<ViiperDevice>(async (resolve, reject) => {
			const socket = new Socket();
			socket.connect(this.port, this.host, async () => {
				try {
					socket.setNoDelay(true);
					
					let wrappedSocket: Socket | any = socket;
					if (this.password) {
						wrappedSocket = await performAuthHandshake(socket, this.password);
					}
					
					const line = `bus/${busId}/${devId}\0`;
					wrappedSocket.write(encoder.encode(line));
					resolve(new ViiperDevice(wrappedSocket));
				} catch (e) {
					socket.end();
					reject(e);
				}
			});
			socket.on('error', reject);
		});
	}

	/**
	 * AddDeviceAndConnect: create a device (JSON request payload) then connect its stream.
	 * Returns the stream device handle and the full Device info response.
	 */
	async addDeviceAndConnect(busId: number, deviceCreateRequest: Types.DeviceCreateRequest): Promise<{ device: ViiperDevice; response: Types.Device }> {
		const resp = await this.busdeviceadd(busId, deviceCreateRequest);
		const devId = resp.devId;
		if (!devId) {
			throw new Error('Device response missing devId');
		}
		const device = await this.connectDevice(busId, devId);
		return { device, response: resp };
	}
}
