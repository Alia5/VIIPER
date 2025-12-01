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

void type_string(viiper::ViiperDevice& stream, const std::string& text) {
    for (char ch : text) {
        auto it = viiper::keyboard::CHAR_TO_KEY.find(static_cast<std::uint8_t>(ch));
        if (it == viiper::keyboard::CHAR_TO_KEY.end()) continue;
        std::uint8_t key = it->second;

        std::uint8_t mods = 0;
        if (viiper::keyboard::SHIFT_CHARS.contains(static_cast<std::uint8_t>(ch))) {
            mods = viiper::keyboard::ModLeftShift;
        }

        viiper::keyboard::Input down = {
            .modifiers = mods,
            .keys = {key},
        };
        stream.send(down);
        std::this_thread::sleep_for(std::chrono::milliseconds(100));

        viiper::keyboard::Input up = {
            .modifiers = 0,
            .keys = {},
        };
        stream.send(up);
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
}

void press_key(viiper::ViiperDevice& stream, std::uint8_t key) {
    viiper::keyboard::Input press = {
        .modifiers = 0,
        .keys = {key},
    };
    stream.send(press);
    std::this_thread::sleep_for(std::chrono::milliseconds(100));

    viiper::keyboard::Input release = {
        .modifiers = 0,
        .keys = {},
    };
    stream.send(release);
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
    auto device_result = client.busdeviceadd(bus_id, {.type = "keyboard"});
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

    stream->on_output(viiper::keyboard::OUTPUT_SIZE, [](const std::uint8_t* data, std::size_t len) {
        if (len < viiper::keyboard::OUTPUT_SIZE) return;
        auto result = viiper::keyboard::Output::from_bytes(data, len);
        if (result.is_error()) return;
        auto& leds = result.value();
        bool num_lock = (leds.leds & 0x01) != 0;
        bool caps_lock = (leds.leds & 0x02) != 0;
        bool scroll_lock = (leds.leds & 0x04) != 0;
        bool compose = (leds.leds & 0x08) != 0;
        bool kana = (leds.leds & 0x10) != 0;
        std::cout << "← LEDs: Num=" << num_lock << " Caps=" << caps_lock
                  << " Scroll=" << scroll_lock << " Compose=" << compose
                  << " Kana=" << kana << "\n";
    });

    std::cout << "Every 5s: type 'Hello!' + Enter. Press Ctrl+C to stop.\n";

    // Type "Hello!" + Enter every 5 seconds
    while (running && stream->is_connected()) {
        type_string(*stream, "Hello!");
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
        press_key(*stream, viiper::keyboard::KeyEnter);

        std::cout << "→ Typed: Hello!\n";
        std::this_thread::sleep_for(std::chrono::seconds(5));
    }

    // Cleanup
    stream->stop();
    client.busdeviceremove(device_info.busid, device_info.devid);
    if (created_bus) {
        client.busremove(bus_id);
    }

    return 0;
}
