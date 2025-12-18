package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const socketTemplate = `// Auto-generated VIIPER C++ Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

#pragma once

#include "../error.hpp"
#include <string>
#include <cstdint>
#include <vector>
#include <memory>
#include <cstring>
#include <mutex>
#include <chrono>

#ifdef _WIN32
#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif
#include <winsock2.h>
#include <ws2tcpip.h>
#pragma comment(lib, "Ws2_32.lib")
#else
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/tcp.h>
#include <netdb.h>
#include <unistd.h>
#include <fcntl.h>
#include <poll.h>
#include <errno.h>
#endif

namespace viiper {
namespace detail {

// ============================================================================
// Windows Socket Initialization
// ============================================================================

#ifdef _WIN32
class WsaInitializer {
public:
    static Result<void> ensure_initialized() {
        static WsaInitializer instance;
        return instance.init_result_;
    }

private:
    WsaInitializer() {
        WSADATA wsa_data;
        if (WSAStartup(MAKEWORD(2, 2), &wsa_data) != 0) {
            init_result_ = Error("WSAStartup failed");
        }
    }

    ~WsaInitializer() {
        if (init_result_.ok()) {
            WSACleanup();
        }
    }

    WsaInitializer(const WsaInitializer&) = delete;
    WsaInitializer& operator=(const WsaInitializer&) = delete;

    Result<void> init_result_;
};
#endif

// ============================================================================
// Cross-platform socket wrapper (thread-safe)
// ============================================================================

class Socket {
public:
    Socket() = default;
    ~Socket() { close(); }

    Socket(const Socket&) = delete;
    Socket& operator=(const Socket&) = delete;

    Socket(Socket&& other) noexcept {
        std::scoped_lock lock(other.send_mutex_, other.recv_mutex_);
        fd_ = other.fd_;
        timeout_ms_ = other.timeout_ms_;
        other.fd_ = invalid_socket();
    }

    Socket& operator=(Socket&& other) noexcept {
        if (this != &other) {
            std::scoped_lock lock(send_mutex_, recv_mutex_, other.send_mutex_, other.recv_mutex_);
            close_internal();
            fd_ = other.fd_;
            timeout_ms_ = other.timeout_ms_;
            other.fd_ = invalid_socket();
        }
        return *this;
    }

    /// Set socket timeout for send/recv operations. Pass 0 to disable timeout.
    Result<void> set_timeout(std::chrono::milliseconds timeout) {
        std::scoped_lock lock(send_mutex_, recv_mutex_);
        timeout_ms_ = static_cast<int>(timeout.count());

        if (!is_valid_internal()) {
            return Result<void>(); // Will apply on next connect
        }

        return apply_timeout_internal();
    }

    Result<void> connect(const std::string& host, std::uint16_t port) {
        std::scoped_lock lock(send_mutex_, recv_mutex_);

#ifdef _WIN32
        auto init_result = WsaInitializer::ensure_initialized();
        if (init_result.is_error()) return init_result.error();
#endif

        addrinfo hints{};
        hints.ai_family = AF_UNSPEC;
        hints.ai_socktype = SOCK_STREAM;
        hints.ai_protocol = IPPROTO_TCP;

        addrinfo* result = nullptr;
        const std::string port_str = std::to_string(port);

        if (::getaddrinfo(host.c_str(), port_str.c_str(), &hints, &result) != 0) {
            return Error("failed to resolve host: " + host);
        }

        std::unique_ptr<addrinfo, decltype(&::freeaddrinfo)> result_guard(result, ::freeaddrinfo);

        for (addrinfo* ptr = result; ptr != nullptr; ptr = ptr->ai_next) {
            auto sock = ::socket(ptr->ai_family, ptr->ai_socktype, ptr->ai_protocol);
            if (sock == invalid_socket()) continue;

            if (::connect(sock, ptr->ai_addr, static_cast<int>(ptr->ai_addrlen)) == 0) {
                fd_ = sock;
                
                int flag = 1;
                ::setsockopt(sock, IPPROTO_TCP, TCP_NODELAY, reinterpret_cast<const char*>(&flag), sizeof(flag));
                
                if (timeout_ms_ > 0) {
                    auto timeout_result = apply_timeout_internal();
                    if (timeout_result.is_error()) {
                        close_internal();
                        return timeout_result.error();
                    }
                }
                return Result<void>();
            }

            close_socket(sock);
        }

        return Error("connection failed: " + host + ":" + port_str);
    }

    Result<void> send(const void* data, std::size_t size) {
        std::lock_guard<std::mutex> lock(send_mutex_);

        if (!is_valid_internal()) {
            return Error("socket not connected");
        }

        std::size_t sent = 0;
        const auto* ptr = static_cast<const char*>(data);

        while (sent < size) {
            auto result = ::send(fd_, ptr + sent, static_cast<int>(size - sent), 0);
            if (result <= 0) {
#ifdef _WIN32
                int err = WSAGetLastError();
                if (err == WSAETIMEDOUT) return Error("send timeout");
#else
                if (errno == EAGAIN || errno == EWOULDBLOCK) return Error("send timeout");
#endif
                return Error("send failed");
            }
            sent += static_cast<std::size_t>(result);
        }

        return Result<void>();
    }

    Result<void> send(const std::string& str) {
        return send(str.data(), str.size());
    }

    Result<std::size_t> recv(void* buffer, std::size_t size) {
        std::lock_guard<std::mutex> lock(recv_mutex_);

        if (!is_valid_internal()) {
            return Error("socket not connected");
        }

        auto result = ::recv(fd_, static_cast<char*>(buffer), static_cast<int>(size), 0);
        if (result < 0) {
#ifdef _WIN32
            int err = WSAGetLastError();
            if (err == WSAETIMEDOUT) return Error("receive timeout");
#else
            if (errno == EAGAIN || errno == EWOULDBLOCK) return Error("receive timeout");
#endif
            return Error("receive failed");
        }
        return static_cast<std::size_t>(result);
    }

    Result<void> recv_exact(void* buffer, std::size_t size) {
        std::lock_guard<std::mutex> lock(recv_mutex_);

        if (!is_valid_internal()) {
            return Error("socket not connected");
        }

        std::size_t received = 0;
        auto* ptr = static_cast<char*>(buffer);

        while (received < size) {
            auto result = ::recv(fd_, ptr + received, static_cast<int>(size - received), 0);
            if (result < 0) {
#ifdef _WIN32
                int err = WSAGetLastError();
                if (err == WSAETIMEDOUT) return Error("receive timeout");
#else
                if (errno == EAGAIN || errno == EWOULDBLOCK) return Error("receive timeout");
#endif
                return Error("receive failed");
            }
            if (result == 0) {
                return Error("connection closed");
            }
            received += static_cast<std::size_t>(result);
        }

        return Result<void>();
    }

    Result<std::string> recv_line() {
        std::lock_guard<std::mutex> lock(recv_mutex_);

        if (!is_valid_internal()) {
            return Error("socket not connected");
        }

        std::string line;
        char ch;

        while (true) {
            auto result = ::recv(fd_, &ch, 1, 0);
            if (result < 0) {
#ifdef _WIN32
                int err = WSAGetLastError();
                if (err == WSAETIMEDOUT) return Error("receive timeout");
#else
                if (errno == EAGAIN || errno == EWOULDBLOCK) return Error("receive timeout");
#endif
                return Error("receive failed");
            }
            if (result == 0) {
                break;
            }
            if (ch == '\0' || ch == '\n') {
                break;
            }
            line += ch;
        }

        return line;
    }

    void close() {
        std::scoped_lock lock(send_mutex_, recv_mutex_);
        close_internal();
    }

    /// Force close the socket without locking. Use with caution - only safe
    /// when you need to interrupt blocking operations from another thread.
    void force_close() noexcept {
        if (fd_ != invalid_socket()) {
            close_socket(fd_);
            fd_ = invalid_socket();
        }
    }

    [[nodiscard]] bool is_valid() const noexcept {
        // Just check fd_ without locking - atomic read on most platforms
        return fd_ != invalid_socket();
    }

private:
    void close_internal() {
        if (is_valid_internal()) {
            close_socket(fd_);
            fd_ = invalid_socket();
        }
    }

    [[nodiscard]] bool is_valid_internal() const noexcept {
        return fd_ != invalid_socket();
    }

    Result<void> apply_timeout_internal() {
#ifdef _WIN32
        DWORD tv = static_cast<DWORD>(timeout_ms_);
        if (setsockopt(fd_, SOL_SOCKET, SO_RCVTIMEO, reinterpret_cast<const char*>(&tv), sizeof(tv)) < 0) {
            return Error("failed to set receive timeout");
        }
        if (setsockopt(fd_, SOL_SOCKET, SO_SNDTIMEO, reinterpret_cast<const char*>(&tv), sizeof(tv)) < 0) {
            return Error("failed to set send timeout");
        }
#else
        struct timeval tv;
        tv.tv_sec = timeout_ms_ / 1000;
        tv.tv_usec = (timeout_ms_ % 1000) * 1000;
        if (setsockopt(fd_, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv)) < 0) {
            return Error("failed to set receive timeout");
        }
        if (setsockopt(fd_, SOL_SOCKET, SO_SNDTIMEO, &tv, sizeof(tv)) < 0) {
            return Error("failed to set send timeout");
        }
#endif
        return Result<void>();
    }

#ifdef _WIN32
    using socket_t = SOCKET;
    static constexpr socket_t invalid_socket() { return INVALID_SOCKET; }

    static void close_socket(socket_t sock) {
        ::closesocket(sock);
    }
#else
    using socket_t = int;
    static constexpr socket_t invalid_socket() { return -1; }

    static void close_socket(socket_t sock) {
        ::close(sock);
    }
#endif

    socket_t fd_ = invalid_socket();
    int timeout_ms_ = 0;
    mutable std::mutex send_mutex_;
    mutable std::mutex recv_mutex_;
};

} // namespace detail
} // namespace viiper
`

func generateSocket(logger *slog.Logger, detailDir string) error {
	logger.Debug("Generating detail/socket.hpp")
	outputFile := filepath.Join(detailDir, "socket.hpp")

	if err := os.WriteFile(outputFile, []byte(socketTemplate), 0644); err != nil {
		return fmt.Errorf("write socket.hpp: %w", err)
	}

	logger.Info("Generated socket.hpp", "file", outputFile)
	return nil
}
