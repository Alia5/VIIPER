package typescript

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"viiper/internal/codegen/common"
	"viiper/internal/codegen/meta"
	"viiper/internal/codegen/scanner"
)

const clientTemplateTS = `{{writeFileHeaderTS}}
import { Socket } from 'net';
import { TextDecoder, TextEncoder } from 'util';
import type * as Types from './types/ManagementDtos';
import { ViiperDevice } from './ViiperDevice';

const encoder = new TextEncoder();
const decoder = new TextDecoder();

/**
 * VIIPER management API client for bus and device control
 */
export class ViiperClient {
  private host: string;
  private port: number;

  constructor(host: string, port: number = 3242) {
    this.host = host;
    this.port = port;
  }
{{range .Routes}}{{if eq .Method "Register"}}
  /**
   * {{.Handler}}: {{.Path}}
   */{{if .ResponseDTO}}
  async {{toCamelCase .Handler}}({{generateMethodParamsTS .}}): Promise<Types.{{.ResponseDTO}}> {{else}}
  async {{toCamelCase .Handler}}({{generateMethodParamsTS .}}): Promise<boolean> {{end}}{
    const path = ` + "`" + `{{.Path}}` + "`" + `{{range $key, $value := .PathParams}}.replace("{{lb}}{{$key}}{{rb}}", String({{toCamelCase $key}})){{end}};
    {{if .Arguments}}const payload = {{range $i, $arg := .Arguments}}{{if $i}} + ' ' + {{end}}String({{toCamelCase $arg.Name}}){{end}};{{else}}const payload: string | null = null;{{end}}
    {{if .ResponseDTO}}return await this.sendRequest<Types.{{.ResponseDTO}}>(path, payload);{{else}}await this.sendRequest<object>(path, payload); return true;{{end}}
  }
{{end}}{{end}}
  private sendRequest<T>(path: string, payload?: string | null): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      const socket = new Socket();
      socket.connect(this.port, this.host, () => {
        let line = path.toLowerCase();
        if (payload && payload.length > 0) line += ' ' + payload;
        line += '\n';
        socket.write(encoder.encode(line));
      });

      let buffer = '';
      socket.on('data', (chunk: Buffer) => {
        buffer += decoder.decode(chunk);
        if (buffer.includes('\n')) {
          const json = buffer.replace(/\n.*/, '');
          try {
            const obj = JSON.parse(json) as T;
            resolve(obj);
          } catch (e) {
            reject(e);
          } finally {
            socket.end();
          }
        }
      });

      socket.on('error', reject);
      socket.on('end', () => {/* noop */});
    });
  }

  async connectDevice(busId: number, devId: string): Promise<ViiperDevice> {
    return new Promise<ViiperDevice>((resolve, reject) => {
      const socket = new Socket();
      socket.connect(this.port, this.host, () => {
        const line = ` + "`" + `bus/${busId}/${devId}\n` + "`" + `;
        socket.write(encoder.encode(line));
        resolve(new ViiperDevice(socket));
      });
      socket.on('error', reject);
    });
  }

  /**
   * AddDeviceAndConnect creates a device on the specified bus and immediately connects to its stream.
   * This is a convenience wrapper that combines busdeviceadd + connectDevice in one call.
   */
  async addDeviceAndConnect(busId: number, deviceType: string): Promise<{ device: ViiperDevice; response: Types.DeviceAddResponse }> {
    const resp = await this.busdeviceadd(busId, deviceType);
    
    // Parse device ID from response (format: "busId-devId")
    const parts = resp.id.split('-');
    if (parts.length < 2) {
      throw new Error(` + "`" + `Invalid device ID format: ${resp.id}` + "`" + `);
    }
    const devId = parts.slice(1).join('-');
    
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
	for _, arg := range route.Arguments {
		tsType := goTypeToTS(arg.Type)
		if arg.Optional {
			params = append(params, fmt.Sprintf("%s?: %s", common.ToCamelCase(arg.Name), tsType))
		} else {
			params = append(params, fmt.Sprintf("%s: %s", common.ToCamelCase(arg.Name), tsType))
		}
	}
	if len(params) == 0 {
		return ""
	}
	return strings.Join(params, ", ")
}
