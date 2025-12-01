package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
)

const deviceTemplate = `// Auto-generated VIIPER C++ SDK
// DO NOT EDIT - This file is generated from the VIIPER server codebase

#pragma once

#include "config.hpp"
#include "error.hpp"
#include "detail/socket.hpp"
#include <string>
#include <memory>
#include <functional>
#include <thread>
#include <atomic>
#include <mutex>
#include <concepts>

namespace viiper {

template<typename T>
concept DeviceInput = requires(T input) {
    { input.to_bytes() } -> std::convertible_to<std::vector<std::uint8_t>>;
};

// ============================================================================
// Device Stream Connection (thread-safe)
// ============================================================================

class ViiperDevice {
public:
    using OutputCallback = std::function<void(const std::uint8_t*, std::size_t)>;
    using DisconnectCallback = std::function<void()>;
    using ErrorCallback = std::function<void(const Error&)>;

    ~ViiperDevice() {
        stop();
    }

    ViiperDevice(const ViiperDevice&) = delete;
    ViiperDevice& operator=(const ViiperDevice&) = delete;
    ViiperDevice(ViiperDevice&&) = delete;
    ViiperDevice& operator=(ViiperDevice&&) = delete;

    // ========================================================================
    // Input (Client -> Device)
    // ========================================================================

    template<DeviceInput T>
    Result<void> send(const T& input) {
        std::lock_guard<std::mutex> lock(send_mutex_);
        auto bytes = input.to_bytes();
        return socket_.send(bytes.data(), bytes.size());
    }

    Result<void> send_raw(const std::uint8_t* data, std::size_t size) {
        std::lock_guard<std::mutex> lock(send_mutex_);
        return socket_.send(data, size);
    }

    // ========================================================================
    // Output (Device -> Client, async)
    // ========================================================================

    Result<void> on_output(std::size_t buffer_size, OutputCallback callback) {
        std::lock_guard<std::mutex> lock(callback_mutex_);

        if (output_thread_.joinable()) {
            return Error("output callback already registered");
        }

        output_callback_ = std::move(callback);
        output_buffer_size_ = buffer_size;
        running_ = true;

        output_thread_ = std::thread([this]() {
            auto buffer = std::make_unique<std::uint8_t[]>(output_buffer_size_);

            while (running_) {
                auto recv_result = socket_.recv(buffer.get(), output_buffer_size_);
                if (recv_result.is_error()) {
                    if (error_callback_) {
                        error_callback_(recv_result.error());
                    }
                    running_ = false;
                    break;
                }

                auto bytes_read = recv_result.value();
                if (bytes_read == 0) {
                    running_ = false;
                    break;
                }

                std::lock_guard<std::mutex> lock(callback_mutex_);
                if (output_callback_) {
                    output_callback_(buffer.get(), bytes_read);
                }
            }

            std::lock_guard<std::mutex> lock(callback_mutex_);
            if (disconnect_callback_) {
                disconnect_callback_();
            }
        });

        return Result<void>();
    }

    void on_disconnect(DisconnectCallback callback) {
        std::lock_guard<std::mutex> lock(callback_mutex_);
        disconnect_callback_ = std::move(callback);
    }

    void on_error(ErrorCallback callback) {
        std::lock_guard<std::mutex> lock(callback_mutex_);
        error_callback_ = std::move(callback);
    }

    void stop() {
        running_ = false;
        socket_.force_close();
        if (output_thread_.joinable()) {
            output_thread_.join();
        }
    }

    [[nodiscard]] bool is_connected() const noexcept {
        return running_.load() && socket_.is_valid();
    }

    explicit ViiperDevice(detail::Socket socket)
        : socket_(std::move(socket)), running_(false), output_buffer_size_(0) {}

private:
    detail::Socket socket_;
    std::atomic<bool> running_;
    std::size_t output_buffer_size_;
    OutputCallback output_callback_;
    DisconnectCallback disconnect_callback_;
    ErrorCallback error_callback_;
    std::thread output_thread_;
    std::mutex send_mutex_;
    std::mutex callback_mutex_;
};

} // namespace viiper
`

func generateDevice(logger *slog.Logger, includeDir string, md *meta.Metadata) error {
	logger.Debug("Generating device.hpp")
	outputFile := filepath.Join(includeDir, "device.hpp")

	if err := os.WriteFile(outputFile, []byte(deviceTemplate), 0644); err != nil {
		return fmt.Errorf("write device.hpp: %w", err)
	}

	logger.Info("Generated device.hpp", "file", outputFile)
	return nil
}
