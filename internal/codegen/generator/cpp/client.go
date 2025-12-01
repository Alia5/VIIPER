package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const clientTemplate = `{{.Header}}
#pragma once

#include "config.hpp"
#include "error.hpp"
#include "types.hpp"
#include "device.hpp"
#include "detail/socket.hpp"
#include "detail/json.hpp"
#include <string>
#include <memory>
#include <sstream>
#include <mutex>

namespace viiper {

// ============================================================================
// VIIPER Management API Client (thread-safe)
// ============================================================================

class ViiperClient {
public:
    ViiperClient(std::string host, std::uint16_t port = 3242)
        : host_(std::move(host)), port_(port) {}

    ~ViiperClient() = default;

    ViiperClient(const ViiperClient&) = delete;
    ViiperClient& operator=(const ViiperClient&) = delete;
    ViiperClient(ViiperClient&&) = delete;
    ViiperClient& operator=(ViiperClient&&) = delete;

    [[nodiscard]] const std::string& host() const noexcept { return host_; }
    [[nodiscard]] std::uint16_t port() const noexcept { return port_; }

    // ========================================================================
    // Management API Methods (all return Result<T>)
    // ========================================================================
{{range .Routes}}{{if eq .Method "Register"}}
    // {{.Handler}}: {{.Path}}
    Result<{{responseCppType .ResponseDTO}}{{if eq (responseCppType .ResponseDTO) ""}}void{{end}}> {{camelcase .Handler}}({{$params := pathParams .Path}}{{range $i, $p := $params}}{{if $i}}, {{end}}{{pathParamType $p}} {{$p}}{{end}}{{$payloadType := payloadCppType .Payload}}{{if ne $payloadType ""}}{{if $params}}, {{end}}{{$payloadType}} payload{{end}}) {
        {{$path := .Path}}{{if $params}}std::string path = format_path("{{$path}}", { {{range $i, $p := $params}}{{if $i}}, {{end}}{ "{{$p}}", {{formatPathParamValue $p}} }{{end}} });{{else}}const std::string path = "{{$path}}";{{end}}
        {{if eq .Payload.Kind "json"}}const std::string payload_str = payload.to_json().dump();{{else if eq .Payload.Kind "numeric"}}const std::string payload_str = {{if .Payload.Required}}std::to_string(payload){{else}}payload.has_value() ? std::to_string(*payload) : ""{{end}};{{else if eq .Payload.Kind "string"}}const std::string& payload_str = payload;{{else}}const std::string payload_str;{{end}}
        auto response = do_request(path, payload_str);
        if (response.is_error()) return response.error();
        {{if .ResponseDTO}}return {{responseCppType .ResponseDTO}}::from_json(response.value());{{else}}return Result<void>();{{end}}
    }
{{end}}{{end}}

    // ========================================================================
    // Device Stream Connection
    // ========================================================================

    /// Connect to an existing device's stream for sending input and receiving output
    [[nodiscard]] Result<std::unique_ptr<ViiperDevice>> connectDevice(
        std::uint32_t bus_id,
        const std::string& dev_id
    ) {
        detail::Socket socket;
        auto conn_result = socket.connect(host_, port_);
        if (conn_result.is_error()) return conn_result.error();

        std::string handshake = "bus/" + std::to_string(bus_id) + "/" + dev_id + '\0';
        auto send_result = socket.send(handshake);
        if (send_result.is_error()) return send_result.error();

        return std::unique_ptr<ViiperDevice>(new ViiperDevice(std::move(socket)));
    }

    /// Create a device and connect to its stream in one step
    [[nodiscard]] Result<std::pair<Device, std::unique_ptr<ViiperDevice>>> addDeviceAndConnect(
        std::uint32_t bus_id,
        const Devicecreaterequest& request
    ) {
        auto device_result = busdeviceadd(bus_id, request);
        if (device_result.is_error()) return device_result.error();

        auto& device_info = device_result.value();
        auto connect_result = connectDevice(bus_id, device_info.devid);
        if (connect_result.is_error()) return connect_result.error();

        return std::make_pair(std::move(device_info), std::move(connect_result.value()));
    }

private:
    Result<json_type> do_request(const std::string& path, const std::string& payload) {
        std::lock_guard<std::mutex> lock(request_mutex_);

        detail::Socket socket;
        auto connect_result = socket.connect(host_, port_);
        if (connect_result.is_error()) return connect_result.error();

        std::string request = path;
        if (!payload.empty()) {
            request += " " + payload;
        }
        request += '\0';

        auto send_result = socket.send(request);
        if (send_result.is_error()) return send_result.error();

        auto recv_result = socket.recv_line();
        if (recv_result.is_error()) return recv_result.error();

        return detail::parse_json_response(recv_result.value());
    }

    static std::string format_path(const std::string& pattern,
                                    std::initializer_list<std::pair<std::string, std::string>> params) {
        std::string result = pattern;
        for (const auto& [name, value] : params) {
            std::string placeholder = "{" + name + "}";
            std::size_t pos = result.find(placeholder);
            if (pos != std::string::npos) {
                result.replace(pos, placeholder.length(), value);
            }
        }
        return result;
    }

    std::string host_;
    std::uint16_t port_;
    mutable std::mutex request_mutex_;
};

} // namespace viiper
`

func generateClient(logger *slog.Logger, includeDir string, md *meta.Metadata) error {
	logger.Debug("Generating client.hpp")
	outputFile := filepath.Join(includeDir, "client.hpp")

	tmpl := template.Must(template.New("client").Funcs(tplFuncs(md)).Parse(clientTemplate))

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create client.hpp: %w", err)
	}
	defer f.Close()

	data := struct {
		Header string
		Routes []scanner.RouteInfo
	}{
		Header: writeFileHeader(),
		Routes: md.Routes,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute client template: %w", err)
	}

	logger.Info("Generated client.hpp", "file", outputFile)
	return nil
}
