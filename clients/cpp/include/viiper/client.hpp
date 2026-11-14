

#pragma once

#include "config.hpp"
#include "error.hpp"
#include "types.hpp"
#include "device.hpp"
#include "detail/socket.hpp"
#include "detail/json.hpp"
#include "detail/auth.hpp"
#include <string>
#include <memory>
#include <sstream>
#include <mutex>

namespace viiper
{

    // ============================================================================
    // VIIPER Management API Client (thread-safe)
    // ============================================================================

    class ViiperClient
    {
    public:
        ViiperClient(std::string host, std::uint16_t port = 3242, std::string password = "")
            : host_(std::move(host)), port_(port), password_(std::move(password)) {}

        ~ViiperClient() = default;

        ViiperClient(const ViiperClient &) = delete;
        ViiperClient &operator=(const ViiperClient &) = delete;
        ViiperClient(ViiperClient &&) = delete;
        ViiperClient &operator=(ViiperClient &&) = delete;

        [[nodiscard]] const std::string &host() const noexcept { return host_; }
        [[nodiscard]] std::uint16_t port() const noexcept { return port_; }
        [[nodiscard]] const std::string &password() const noexcept { return password_; }

        // ========================================================================
        // Management API Methods (all return Result<T>)
        // ========================================================================

        // Ping: ping
        Result<Pingresponse> ping()
        {
            const std::string path = "ping";
            const std::string payload_str;
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Pingresponse::from_json(response.value());
        }

        // BusList: bus/list
        Result<Buslistresponse> buslist()
        {
            const std::string path = "bus/list";
            const std::string payload_str;
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Buslistresponse::from_json(response.value());
        }

        // BusCreate: bus/create
        Result<Buscreateresponse> buscreate(std::optional<std::uint32_t> payload)
        {
            const std::string path = "bus/create";
            const std::string payload_str = payload.has_value() ? std::to_string(*payload) : "";
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Buscreateresponse::from_json(response.value());
        }

        // BusRemove: bus/remove
        Result<Busremoveresponse> busremove(std::uint32_t payload)
        {
            const std::string path = "bus/remove";
            const std::string payload_str = std::to_string(payload);
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Busremoveresponse::from_json(response.value());
        }

        // BusDevicesList: bus/{id}/list
        Result<Deviceslistresponse> busdeviceslist(std::uint32_t id)
        {
            std::string path = format_path("bus/{id}/list", {{"id", std::to_string(id)}});
            const std::string payload_str;
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Deviceslistresponse::from_json(response.value());
        }

        // BusDeviceAdd: bus/{id}/add
        Result<Device> busdeviceadd(std::uint32_t id, const Devicecreaterequest &payload)
        {
            std::string path = format_path("bus/{id}/add", {{"id", std::to_string(id)}});
            const std::string payload_str = payload.to_json().dump();
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Device::from_json(response.value());
        }

        // BusDeviceRemove: bus/{id}/remove
        Result<Deviceremoveresponse> busdeviceremove(std::uint32_t id, const std::string &payload)
        {
            std::string path = format_path("bus/{id}/remove", {{"id", std::to_string(id)}});
            const std::string &payload_str = payload;
            auto response = do_request(path, payload_str);
            if (response.is_error())
                return response.error();
            return Deviceremoveresponse::from_json(response.value());
        }

        // ========================================================================
        // Device Stream Connection
        // ========================================================================

        /// Connect to an existing device's stream for sending input and receiving output
        [[nodiscard]] Result<std::unique_ptr<ViiperDevice>> connectDevice(
            std::uint32_t bus_id,
            const std::string &dev_id)
        {
            detail::Socket socket;
            auto conn_result = socket.connect(host_, port_);
            if (conn_result.is_error())
                return conn_result.error();

            std::string handshake = "bus/" + std::to_string(bus_id) + "/" + dev_id + '\0';

            if (!password_.empty())
            {
                auto handshake_result = detail::perform_handshake(std::move(socket), password_);
                if (handshake_result.is_error())
                    return handshake_result.error();

                auto encrypted_socket = std::move(handshake_result.value());
                auto send_result = encrypted_socket->send(handshake);
                if (send_result.is_error())
                    return send_result.error();

                return std::unique_ptr<ViiperDevice>(new ViiperDevice(std::move(encrypted_socket)));
            }
            else
            {
                auto send_result = socket.send(handshake);
                if (send_result.is_error())
                    return send_result.error();

                return std::unique_ptr<ViiperDevice>(new ViiperDevice(std::move(socket)));
            }
        }

        /// Create a device and connect to its stream in one step
        [[nodiscard]] Result<std::pair<Device, std::unique_ptr<ViiperDevice>>> addDeviceAndConnect(
            std::uint32_t bus_id,
            const Devicecreaterequest &request)
        {
            auto device_result = busdeviceadd(bus_id, request);
            if (device_result.is_error())
                return device_result.error();

            auto &device_info = device_result.value();
            auto connect_result = connectDevice(bus_id, device_info.devid);
            if (connect_result.is_error())
                return connect_result.error();

            return std::make_pair(std::move(device_info), std::move(connect_result.value()));
        }

    private:
        Result<json_type> do_request(const std::string &path, const std::string &payload)
        {
            std::lock_guard<std::mutex> lock(request_mutex_);

            detail::Socket socket;
            auto connect_result = socket.connect(host_, port_);
            if (connect_result.is_error())
                return connect_result.error();

            std::string request = path;
            if (!payload.empty())
            {
                request += " " + payload;
            }
            request += '\0';

            if (!password_.empty())
            {
                auto handshake_result = detail::perform_handshake(std::move(socket), password_);
                if (handshake_result.is_error())
                    return handshake_result.error();

                auto encrypted_socket = std::move(handshake_result.value());
                auto send_result = encrypted_socket->send(request);
                if (send_result.is_error())
                    return send_result.error();

                auto recv_result = encrypted_socket->recv_line();
                if (recv_result.is_error())
                    return recv_result.error();

                return detail::parse_json_response(recv_result.value());
            }
            else
            {
                auto send_result = socket.send(request);
                if (send_result.is_error())
                    return send_result.error();

                auto recv_result = socket.recv_line();
                if (recv_result.is_error())
                    return recv_result.error();

                return detail::parse_json_response(recv_result.value());
            }
        }

        static std::string format_path(const std::string &pattern,
                                       std::initializer_list<std::pair<std::string, std::string>> params)
        {
            std::string result = pattern;
            for (const auto &[name, value] : params)
            {
                std::string placeholder = "{" + name + "}";
                std::size_t pos = result.find(placeholder);
                if (pos != std::string::npos)
                {
                    result.replace(pos, placeholder.length(), value);
                }
            }
            return result;
        }

        std::string host_;
        std::uint16_t port_;
        std::string password_;
        mutable std::mutex request_mutex_;
    };

} // namespace viiper
