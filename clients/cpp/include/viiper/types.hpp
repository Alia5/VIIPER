

#pragma once

#include "config.hpp"
#include "detail/json.hpp"
#include <string>
#include <vector>
#include <optional>
#include <cstdint>

namespace viiper
{

    struct Apierror;
    struct Pingresponse;
    struct Buslistresponse;
    struct Buscreateresponse;
    struct Busremoveresponse;
    struct Device;
    struct Deviceslistresponse;
    struct Deviceremoveresponse;
    struct Devicecreaterequest;

    // ============================================================================
    // Management API DTOs
    // ============================================================================

    // APIError
    struct Apierror
    {
        std::int32_t status;
        std::string title;
        std::string detail;

        static Apierror from_json(const json_type &j)
        {
            Apierror result;
            result.status = j.value("status", std::int32_t{});
            result.title = j.value("title", std::string{});
            result.detail = j.value("detail", std::string{});
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            j["status"] = status;
            j["title"] = title;
            j["detail"] = detail;
            return j;
        }
    };

    // PingResponse
    struct Pingresponse
    {
        std::string server;
        std::string version;

        static Pingresponse from_json(const json_type &j)
        {
            Pingresponse result;
            result.server = j.value("server", std::string{});
            result.version = j.value("version", std::string{});
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            j["server"] = server;
            j["version"] = version;
            return j;
        }
    };

    // BusListResponse
    struct Buslistresponse
    {
        std::vector<std::uint32_t> buses;

        static Buslistresponse from_json(const json_type &j)
        {
            Buslistresponse result;
            result.buses = detail::get_array<std::uint32_t>(j, "buses");
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            {
                json_type arr = json_type::array();
                for (const auto &item : buses)
                {
                    arr.push_back(item);
                }
                j["buses"] = std::move(arr);
            }
            return j;
        }
    };

    // BusCreateResponse
    struct Buscreateresponse
    {
        std::uint32_t busid;

        static Buscreateresponse from_json(const json_type &j)
        {
            Buscreateresponse result;
            result.busid = j.value("busId", std::uint32_t{});
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            j["busId"] = busid;
            return j;
        }
    };

    // BusRemoveResponse
    struct Busremoveresponse
    {
        std::uint32_t busid;

        static Busremoveresponse from_json(const json_type &j)
        {
            Busremoveresponse result;
            result.busid = j.value("busId", std::uint32_t{});
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            j["busId"] = busid;
            return j;
        }
    };

    // Device
    struct Device
    {
        std::uint32_t busid;
        std::string devid;
        std::string vid;
        std::string pid;
        std::string type;
        json_type devicespecific;

        static Device from_json(const json_type &j)
        {
            Device result;
            result.busid = j.value("busId", std::uint32_t{});
            result.devid = j.value("devId", std::string{});
            result.vid = j.value("vid", std::string{});
            result.pid = j.value("pid", std::string{});
            result.type = j.value("type", std::string{});
            if (j.contains("deviceSpecific"))
            {
                result.devicespecific = j["deviceSpecific"];
            }
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            j["busId"] = busid;
            j["devId"] = devid;
            j["vid"] = vid;
            j["pid"] = pid;
            j["type"] = type;
            j["deviceSpecific"] = devicespecific;
            return j;
        }
    };

    // DevicesListResponse
    struct Deviceslistresponse
    {
        std::vector<Device> devices;

        static Deviceslistresponse from_json(const json_type &j)
        {
            Deviceslistresponse result;
            result.devices = detail::get_array<Device>(j, "devices");
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            {
                json_type arr = json_type::array();
                for (const auto &item : devices)
                {
                    arr.push_back(item.to_json());
                }
                j["devices"] = std::move(arr);
            }
            return j;
        }
    };

    // DeviceRemoveResponse
    struct Deviceremoveresponse
    {
        std::uint32_t busid;
        std::string devid;

        static Deviceremoveresponse from_json(const json_type &j)
        {
            Deviceremoveresponse result;
            result.busid = j.value("busId", std::uint32_t{});
            result.devid = j.value("devId", std::string{});
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            j["busId"] = busid;
            j["devId"] = devid;
            return j;
        }
    };

    // DeviceCreateRequest
    struct Devicecreaterequest
    {
        std::optional<std::string> type;
        std::optional<std::uint16_t> idvendor;
        std::optional<std::uint16_t> idproduct;
        std::optional<json_type> devicespecific;

        static Devicecreaterequest from_json(const json_type &j)
        {
            Devicecreaterequest result;
            result.type = detail::get_optional_field<std::string>(j, "type");
            result.idvendor = detail::get_optional_field<std::uint16_t>(j, "idVendor");
            result.idproduct = detail::get_optional_field<std::uint16_t>(j, "idProduct");
            if (j.contains("deviceSpecific") && !j["deviceSpecific"].is_null())
            {
                result.devicespecific = j["deviceSpecific"];
            }
            else
            {
                result.devicespecific = std::nullopt;
            }
            return result;
        }

        [[nodiscard]] json_type to_json() const
        {
            json_type j;
            if (type.has_value())
            {
                j["type"] = type.value();
            }
            if (idvendor.has_value())
            {
                j["idVendor"] = idvendor.value();
            }
            if (idproduct.has_value())
            {
                j["idProduct"] = idproduct.value();
            }
            if (devicespecific.has_value())
            {
                j["deviceSpecific"] = devicespecific.value();
            }
            return j;
        }
    };

} // namespace viiper
