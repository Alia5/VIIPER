package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const clientTemplateTS = `{{writeFileHeaderTS}}
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
{{range .Routes}}{{if eq .Method "Register"}}
	/**
	 * {{.Handler}}: {{.Path}}
	 */{{if .ResponseDTO}}
	async {{toCamelCase .Handler}}({{generateMethodParamsTS .}}): Promise<Types.{{.ResponseDTO}}> {{else}}
	async {{toCamelCase .Handler}}({{generateMethodParamsTS .}}): Promise<boolean> {{end}}{
		const path = ` + "`" + `{{.Path}}` + "`" + `{{range $key, $value := .PathParams}}.replace("{{lb}}{{$key}}{{rb}}", String({{toCamelCase $key}})){{end}};
		{{if eq .Payload.Kind "none"}}const payload: string = '';{{else if eq .Payload.Kind "json"}}const payload: string = JSON.stringify({{payloadParamNameTS .}});{{else if eq .Payload.Kind "numeric"}}const payload: string = {{payloadParamNameTS .}} !== undefined && {{payloadParamNameTS .}} !== null ? String({{payloadParamNameTS .}}) : '';{{else if eq .Payload.Kind "string"}}const payload: string = {{payloadParamNameTS .}} ? String({{payloadParamNameTS .}}) : '';{{end}}
		{{if .ResponseDTO}}return await this.sendRequest<Types.{{.ResponseDTO}}>(path, payload);{{else}}await this.sendRequest<object>(path, payload); return true;{{end}}
	}
{{end}}{{end}}
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
					
					const line = ` + "`" + `bus/${busId}/${devId}\0` + "`" + `;
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
`

func generateClient(logger *slog.Logger, srcDir string, md *meta.Metadata) error {
	logger.Debug("Generating ViiperClient.ts management API")
	outputFile := filepath.Join(srcDir, "ViiperClient.ts")
	funcMap := template.FuncMap{
		"writeFileHeaderTS":      writeFileHeaderTS,
		"toCamelCase":            common.ToCamelCase,
		"generateMethodParamsTS": generateMethodParamsTS,
		"payloadParamNameTS":     payloadParamNameTS,
		"lb":                     func() string { return "{" },
		"rb":                     func() string { return "}" },
	}
	tmpl, err := template.New("clientTS").Funcs(funcMap).Parse(clientTemplateTS)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	data := struct{ Routes []scanner.RouteInfo }{Routes: md.Routes}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	logger.Info("Generated ViiperClient.ts", "file", outputFile)
	return nil
}

func generateMethodParamsTS(route scanner.RouteInfo) string {
	var params []string
	for key := range route.PathParams {
		params = append(params, fmt.Sprintf("%s: number", common.ToCamelCase(key)))
	}
	switch route.Payload.Kind {
	case scanner.PayloadJSON:
		name := payloadParamNameTS(route)
		ptype := "any"
		if route.Payload.RawType != "" {
			ptype = fmt.Sprintf("Types.%s", route.Payload.RawType)
		}
		params = append(params, fmt.Sprintf("%s: %s", name, ptype))
	case scanner.PayloadNumeric:
		name := payloadParamNameTS(route)
		if route.Payload.Required {
			params = append(params, fmt.Sprintf("%s: number", name))
		} else {
			params = append(params, fmt.Sprintf("%s?: number", name))
		}
	case scanner.PayloadString:
		name := payloadParamNameTS(route)
		if route.Payload.Required {
			params = append(params, fmt.Sprintf("%s: string", name))
		} else {
			params = append(params, fmt.Sprintf("%s?: string", name))
		}
	}
	return strings.Join(params, ", ")
}

func payloadParamNameTS(route scanner.RouteInfo) string {
	if route.Payload.Kind == scanner.PayloadNone {
		return ""
	}
	hint := route.Payload.ParserHint
	if hint == "" {
		return "payload"
	}
	switch route.Payload.Kind {
	case scanner.PayloadNumeric:
		if strings.Contains(strings.ToLower(hint), "id") || strings.HasPrefix(hint, "uint") || strings.HasPrefix(hint, "int") {
			return "id"
		}
		return "value"
	case scanner.PayloadJSON:
		if route.Payload.RawType != "" {
			return common.ToCamelCase(route.Payload.RawType)
		}
		return "request"
	case scanner.PayloadString:
		return "value"
	}
	return "payload"
}
