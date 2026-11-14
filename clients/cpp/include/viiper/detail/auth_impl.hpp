// Auto-generated VIIPER C++ Client Library - Authentication Implementation
// DO NOT EDIT - This file is generated from the VIIPER server codebase

#pragma once

#include <cstdint>
#include <cstring>
#include <algorithm>
#include <random>
#include <vector>
#include <array>
#include <memory>
#include <openssl/evp.h>
#include <openssl/hmac.h>
#include <openssl/rand.h>
#include <openssl/sha.h>
#include "socket.hpp"
#include "../error.hpp"

namespace viiper {
namespace detail {

// ============================================================================
// Crypto Constants
// ============================================================================

constexpr const char* HANDSHAKE_MAGIC = "eVI1\x00";
constexpr size_t NONCE_SIZE = 32;
constexpr const char* AUTH_CONTEXT = "VIIPER-Auth-v1";
constexpr const char* SESSION_CONTEXT = "VIIPER-Session-v1";
constexpr const char* PBKDF2_SALT = "VIIPER-Key-v1";
constexpr uint32_t PBKDF2_ITERATIONS = 100000;

// ============================================================================
// OpenSSL-based Crypto Utilities
// ============================================================================

// PBKDF2-HMAC-SHA256 using OpenSSL
inline void pbkdf2_hmac_sha256(const uint8_t* password, size_t password_len,
                                const uint8_t* salt, size_t salt_len,
                                uint32_t iterations, uint8_t* out, size_t out_len) {
    PKCS5_PBKDF2_HMAC(reinterpret_cast<const char*>(password), static_cast<int>(password_len),
                      salt, static_cast<int>(salt_len),
                      static_cast<int>(iterations),
                      EVP_sha256(),
                      static_cast<int>(out_len), out);
}

// HMAC-SHA256 using OpenSSL
inline void hmac_sha256(const uint8_t* key, size_t key_len, const uint8_t* data, size_t data_len, uint8_t* out) {
    unsigned int len = 32;
    HMAC(EVP_sha256(), key, static_cast<int>(key_len), data, data_len, out, &len);
}

// SHA-256 using OpenSSL
inline void sha256(const uint8_t* data, size_t len, uint8_t* out) {
    EVP_MD_CTX* ctx = EVP_MD_CTX_new();
    EVP_DigestInit_ex(ctx, EVP_sha256(), nullptr);
    EVP_DigestUpdate(ctx, data, len);
    EVP_DigestFinal_ex(ctx, out, nullptr);
    EVP_MD_CTX_free(ctx);
}

// ============================================================================
// ChaCha20-Poly1305 AEAD using OpenSSL
// ============================================================================

class ChaCha20Poly1305 {
private:
    uint8_t key_[32];

public:
    ChaCha20Poly1305(const uint8_t key[32]) {
        std::memcpy(key_, key, 32);
    }

    void encrypt(const uint8_t nonce[12], const uint8_t* plaintext, size_t pt_len,
                 uint8_t* ciphertext, uint8_t tag[16]) {
        EVP_CIPHER_CTX* ctx = EVP_CIPHER_CTX_new();
        EVP_EncryptInit_ex(ctx, EVP_chacha20_poly1305(), nullptr, key_, nonce);
        
        int len;
        EVP_EncryptUpdate(ctx, ciphertext, &len, plaintext, static_cast<int>(pt_len));
        
        int ciphertext_len = len;
        EVP_EncryptFinal_ex(ctx, ciphertext + len, &len);
        ciphertext_len += len;
        
        EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_AEAD_GET_TAG, 16, tag);
        EVP_CIPHER_CTX_free(ctx);
    }

    bool decrypt(const uint8_t nonce[12], const uint8_t* ciphertext, size_t ct_len,
                 const uint8_t tag[16], uint8_t* plaintext) {
        EVP_CIPHER_CTX* ctx = EVP_CIPHER_CTX_new();
        EVP_DecryptInit_ex(ctx, EVP_chacha20_poly1305(), nullptr, key_, nonce);
        
        int len;
        EVP_DecryptUpdate(ctx, plaintext, &len, ciphertext, static_cast<int>(ct_len));
        
        EVP_CIPHER_CTX_ctrl(ctx, EVP_CTRL_AEAD_SET_TAG, 16, const_cast<uint8_t*>(tag));
        
        int ret = EVP_DecryptFinal_ex(ctx, plaintext + len, &len);
        EVP_CIPHER_CTX_free(ctx);
        
        return ret > 0;
    }
};

// ============================================================================
// Encrypted Socket Wrapper
// ============================================================================

class EncryptedSocket {
private:
    Socket socket_;
    ChaCha20Poly1305 cipher_;
    uint64_t send_counter_ = 0;
    std::vector<uint8_t> recv_buffer_;

public:
    EncryptedSocket(Socket&& socket, const std::array<uint8_t, 32>& session_key)
        : socket_(std::move(socket)), cipher_(session_key.data()) {}


    Result<void> send(const std::string& data) {
        return send(reinterpret_cast<const uint8_t*>(data.data()), data.size());
    }

    Result<void> send(const uint8_t* data, size_t size) {
        uint8_t nonce[12] = {0};
        for (int i = 0; i < 8; ++i) {
            nonce[4 + i] = (send_counter_ >> (56 - i * 8)) & 0xff;
        }
        send_counter_++;

        std::vector<uint8_t> ciphertext(size);
        uint8_t tag[16];
        cipher_.encrypt(nonce, data, size, ciphertext.data(), tag);

        uint32_t packet_len = static_cast<uint32_t>(12 + ciphertext.size() + 16);
        uint8_t len_bytes[4] = {
            static_cast<uint8_t>((packet_len >> 24) & 0xff),
            static_cast<uint8_t>((packet_len >> 16) & 0xff),
            static_cast<uint8_t>((packet_len >> 8) & 0xff),
            static_cast<uint8_t>(packet_len & 0xff)
        };

        std::string packet;
        packet.append(reinterpret_cast<char*>(len_bytes), 4);
        packet.append(reinterpret_cast<char*>(nonce), 12);
        packet.append(reinterpret_cast<char*>(ciphertext.data()), ciphertext.size());
        packet.append(reinterpret_cast<char*>(tag), 16);

        return socket_.send(packet);
    }

    Result<size_t> recv(uint8_t* buffer, size_t size) {
        std::vector<uint8_t> len_buf(4);
        auto read_result = socket_.recv_exact(len_buf.data(), 4);
        if (read_result.is_error()) {
            if (read_result.error().message == "connection closed") {
                return 0; // Return 0 bytes on EOF
            }
            return read_result.error();
        }

        uint32_t packet_len = (len_buf[0] << 24) | (len_buf[1] << 16) | 
                             (len_buf[2] << 8) | len_buf[3];

        if (packet_len > 2 * 1024 * 1024) {
            return Error("Packet too large");
        }

        std::vector<uint8_t> packet(packet_len);
        read_result = socket_.recv_exact(packet.data(), packet_len);
        if (read_result.is_error()) return read_result.error();

        if (packet_len < 28) return Error("Packet too small");

        const uint8_t* nonce = packet.data();
        const uint8_t* ciphertext = packet.data() + 12;
        size_t ct_len = packet_len - 12 - 16;
        const uint8_t* tag = packet.data() + 12 + ct_len;

        std::vector<uint8_t> plaintext(ct_len);
        if (!cipher_.decrypt(nonce, ciphertext, ct_len, tag, plaintext.data())) {
            return Error("Decryption failed");
        }

        size_t to_copy = (ct_len < size) ? ct_len : size;
        std::memcpy(buffer, plaintext.data(), to_copy);
        return to_copy;
    }

    Result<std::string> recv_line() {
        std::vector<uint8_t> len_buf(4);
        auto read_result = socket_.recv_exact(len_buf.data(), 4);
        if (read_result.is_error()) {
            return read_result.error();
        }

        uint32_t packet_len = (len_buf[0] << 24) | (len_buf[1] << 16) | 
                             (len_buf[2] << 8) | len_buf[3];

        if (packet_len > 2 * 1024 * 1024) {
            return Error("Packet too large");
        }

        std::vector<uint8_t> packet(packet_len);
        read_result = socket_.recv_exact(packet.data(), packet_len);
        if (read_result.is_error()) return read_result.error();

        if (packet_len < 28) return Error("Packet too small");

        const uint8_t* nonce = packet.data();
        const uint8_t* ciphertext = packet.data() + 12;
        size_t ct_len = packet_len - 12 - 16;
        const uint8_t* tag = packet.data() + 12 + ct_len;

        std::vector<uint8_t> plaintext(ct_len);
        if (!cipher_.decrypt(nonce, ciphertext, ct_len, tag, plaintext.data())) {
            return Error("Decryption failed");
        }

        return std::string(reinterpret_cast<char*>(plaintext.data()), plaintext.size());
    }

    Socket& get_socket() { return socket_; }
    
    bool is_valid() const { return socket_.is_valid(); }
    
    void force_close() { socket_.force_close(); }
};

// ============================================================================
// Main Handshake Function
// ============================================================================

inline Result<std::unique_ptr<EncryptedSocket>> perform_handshake(Socket&& socket, const std::string& password) {
    if (password.empty()) {
        return Error("Password cannot be empty");
    }

    std::array<uint8_t, 32> key;
    pbkdf2_hmac_sha256(
        reinterpret_cast<const uint8_t*>(password.data()), password.size(),
        reinterpret_cast<const uint8_t*>(PBKDF2_SALT), std::strlen(PBKDF2_SALT),
        PBKDF2_ITERATIONS, key.data(), 32
    );

    std::array<uint8_t, NONCE_SIZE> client_nonce;
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 255);
    for (auto& byte : client_nonce) byte = static_cast<uint8_t>(dis(gen));

    std::array<uint8_t, 32> auth_tag;
    std::vector<uint8_t> auth_data;
    auth_data.insert(auth_data.end(), AUTH_CONTEXT, AUTH_CONTEXT + std::strlen(AUTH_CONTEXT));
    auth_data.insert(auth_data.end(), client_nonce.begin(), client_nonce.end());
    hmac_sha256(key.data(), 32, auth_data.data(), auth_data.size(), auth_tag.data());

    std::string handshake;
    handshake.append(HANDSHAKE_MAGIC, 5);
    handshake.append(reinterpret_cast<char*>(client_nonce.data()), NONCE_SIZE);
    handshake.append(reinterpret_cast<char*>(auth_tag.data()), 32);

    auto send_result = socket.send(handshake);
    if (send_result.is_error()) return send_result.error();

    std::vector<uint8_t> response(3 + NONCE_SIZE);
    auto recv_result = socket.recv_exact(response.data(), response.size());
    if (recv_result.is_error()) return recv_result.error();

    if (response[0] != 'O' || response[1] != 'K' || response[2] != '\0') {
        // Try to read error message
        auto error_data = socket.recv_line();
        if (error_data.is_error()) {
            return Error("Invalid handshake response");
        }
        return Error("Authentication failed: " + error_data.value());
    }

    std::array<uint8_t, NONCE_SIZE> server_nonce;
    std::memcpy(server_nonce.data(), response.data() + 3, NONCE_SIZE);

    std::vector<uint8_t> session_data;
    session_data.insert(session_data.end(), key.begin(), key.end());
    session_data.insert(session_data.end(), server_nonce.begin(), server_nonce.end());
    session_data.insert(session_data.end(), client_nonce.begin(), client_nonce.end());
    session_data.insert(session_data.end(), SESSION_CONTEXT, SESSION_CONTEXT + std::strlen(SESSION_CONTEXT));
    
    std::array<uint8_t, 32> session_key;
    sha256(session_data.data(), session_data.size(), session_key.data());

    auto encrypted = std::make_unique<EncryptedSocket>(std::move(socket), session_key);
    return encrypted;
}

} // namespace detail
} // namespace viiper
