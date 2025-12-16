#define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
#define VIIPER_JSON_NAMESPACE nlohmann
#define VIIPER_JSON_TYPE json

#include <viiper/viiper.hpp>
#include <iostream>
#include <thread>
#include <chrono>
#include <csignal>
#include <atomic>

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
    auto device_result = client.busdeviceadd(bus_id, {.type = "mouse"});
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

    std::cout << "Every 3s: move diagonally by 50px (X and Y), then click and scroll. Press Ctrl+C to stop.\n";

    // Send a short movement once every 3 seconds for easy local testing.
    // Followed by a short click and a single scroll notch.
    int dir = 1;
    constexpr int step = 50; // move diagonally by 50 px in X and Y

    while (running && stream->is_connected()) {
        // Move diagonally: (+step,+step) then (-step,-step) next tick
        std::int16_t dx = static_cast<std::int16_t>(step * dir);
        std::int16_t dy = static_cast<std::int16_t>(step * dir);
        dir *= -1;

        // One-shot movement report (diagonal)
        auto send_result = stream->send(viiper::mouse::Input{
            .buttons = 0,
            .dx = dx,
            .dy = dy,
            .wheel = 0,
            .pan = 0,
        });
        if (send_result.is_error()) {
            std::cerr << "Write error: " << send_result.error().to_string() << "\n";
            break;
        }
        std::cout << "→ Moved mouse dx=" << static_cast<int>(dx) << " dy=" << static_cast<int>(dy) << "\n";

        // Zero state shortly after to keep movement one-shot (harmless safety)
        std::this_thread::sleep_for(std::chrono::milliseconds(30));
        stream->send(viiper::mouse::Input{.buttons = 0, .dx = 0, .dy = 0, .wheel = 0, .pan = 0});

        // Simulate a short left click: press then release
        std::this_thread::sleep_for(std::chrono::milliseconds(50));
        stream->send(viiper::mouse::Input{
            .buttons = viiper::mouse::Btn_Left,
            .dx = 0, .dy = 0, .wheel = 0, .pan = 0,
        });
        std::this_thread::sleep_for(std::chrono::milliseconds(60));
        stream->send(viiper::mouse::Input{.buttons = 0, .dx = 0, .dy = 0, .wheel = 0, .pan = 0});
        std::cout << "→ Clicked (left)\n";

        // Simulate a short scroll: one notch upwards
        std::this_thread::sleep_for(std::chrono::milliseconds(50));
        stream->send(viiper::mouse::Input{.buttons = 0, .dx = 0, .dy = 0, .wheel = 1, .pan = 0});
        std::this_thread::sleep_for(std::chrono::milliseconds(30));
        stream->send(viiper::mouse::Input{.buttons = 0, .dx = 0, .dy = 0, .wheel = 0, .pan = 0});
        std::cout << "→ Scrolled (wheel=+1)\n";

        std::this_thread::sleep_for(std::chrono::seconds(3));
    }

    // Cleanup
    stream->stop();
    client.busdeviceremove(device_info.busid, device_info.devid);
    if (created_bus) {
        client.busremove(bus_id);
    }

    return 0;
}
