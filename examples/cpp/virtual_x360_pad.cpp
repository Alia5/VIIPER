#define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
#define VIIPER_JSON_NAMESPACE nlohmann
#define VIIPER_JSON_TYPE json

#include <viiper/viiper.hpp>
#include <iostream>
#include <thread>
#include <chrono>
#include <csignal>
#include <atomic>
#include <cmath>

std::atomic<bool> running{true};

void signal_handler(int) {
    running = false;
}

int main(int argc, char** argv) {
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <api_addr>\n";
        std::cerr << "Example: " << argv[0] << " localhost:3242\n";
        return 1;
    }

    std::signal(SIGINT, signal_handler);
    std::signal(SIGTERM, signal_handler);

    const std::string addr = argv[1];
    const auto colon_pos = addr.find(':');
    const std::string host = addr.substr(0, colon_pos);
    const std::uint16_t port = colon_pos != std::string::npos
        ? static_cast<std::uint16_t>(std::stoul(addr.substr(colon_pos + 1)))
        : 3242;

    viiper::ViiperClient client(host, port);

    // Find or create a bus
    std::uint32_t bus_id;
    bool created_bus = false;

    auto buses_result = client.buslist();
    if (buses_result.is_error()) {
        std::cerr << "BusList error: " << buses_result.error().to_string() << "\n";
        return 1;
    }

    if (buses_result.value().buses.empty()) {
        auto create_result = client.buscreate(std::nullopt);
        if (create_result.is_error()) {
            std::cerr << "BusCreate failed: " << create_result.error().to_string() << "\n";
            return 1;
        }
        bus_id = create_result.value().busid;
        created_bus = true;
        std::cout << "Created bus " << bus_id << "\n";
    } else {
        bus_id = buses_result.value().buses[0];
        std::cout << "Using existing bus " << bus_id << "\n";
    }

    // Add device
    auto device_result = client.busdeviceadd(bus_id, {.type = "xbox360"});
    if (device_result.is_error()) {
        std::cerr << "AddDevice error: " << device_result.error().to_string() << "\n";
        if (created_bus) {
            client.busremove(bus_id);
        }
        return 1;
    }
    auto device_info = std::move(device_result.value());

    // Connect to device stream
    auto stream_result = client.connectDevice(device_info.busid, device_info.devid);
    if (stream_result.is_error()) {
        std::cerr << "ConnectDevice error: " << stream_result.error().to_string() << "\n";
        client.busdeviceremove(device_info.busid, device_info.devid);
        if (created_bus) {
            client.busremove(bus_id);
        }
        return 1;
    }
    auto stream = std::move(stream_result.value());

    std::cout << "Created and connected to device " << device_info.devid
              << " on bus " << device_info.busid << "\n";

    stream->on_disconnect([]() {
        std::cerr << "Device disconnected by server\n";
        std::exit(0);
    });

    stream->on_output(viiper::xbox360::OUTPUT_SIZE, [](const std::uint8_t* data, std::size_t len) {
        if (len < viiper::xbox360::OUTPUT_SIZE) return;
        auto result = viiper::xbox360::Output::from_bytes(data, len);
        if (result.is_error()) return;
        auto& rumble = result.value();
        std::cout << "← Rumble: Left=" << static_cast<int>(rumble.left)
                  << ", Right=" << static_cast<int>(rumble.right) << "\n";
    });

    // Send controller inputs at 60fps (16ms intervals)
    std::uint64_t frame = 0;

    while (running && stream->is_connected()) {
        ++frame;

        std::uint16_t buttons;
        switch ((frame / 60) % 4) {
            case 0: buttons = viiper::xbox360::ButtonA; break;
            case 1: buttons = viiper::xbox360::ButtonB; break;
            case 2: buttons = viiper::xbox360::ButtonX; break;
            default: buttons = viiper::xbox360::ButtonY; break;
        }

        viiper::xbox360::Input state = {
            .buttons = buttons,
            .lt = static_cast<std::uint8_t>((frame * 2) % 256),
            .rt = static_cast<std::uint8_t>((frame * 3) % 256),
            .lx = static_cast<std::int16_t>(20000.0 * 0.7071),
            .ly = static_cast<std::int16_t>(20000.0 * 0.7071),
            .rx = 0,
            .ry = 0,
        };

        auto send_result = stream->send(state);
        if (send_result.is_error()) {
            std::cerr << "Write error: " << send_result.error().to_string() << "\n";
            break;
        }

        if (frame % 60 == 0) {
            std::cout << "→ Sent input (frame " << frame << "): buttons=0x"
                      << std::hex << state.buttons << std::dec
                      << ", LT=" << static_cast<int>(state.lt)
                      << ", RT=" << static_cast<int>(state.rt) << "\n";
        }

        std::this_thread::sleep_for(std::chrono::milliseconds(16));
    }

    // Cleanup
    stream->stop();
    client.busdeviceremove(device_info.busid, device_info.devid);
    if (created_bus) {
        client.busremove(bus_id);
    }

    return 0;
}
