

#pragma once

#include "socket.hpp"
#include "../error.hpp"
#include <string>
#include <vector>
#include <array>
#include <cstring>
#include <random>

#if __has_include(<openssl/evp.h>)
#define VIIPER_HAS_OPENSSL 1
#include <openssl/evp.h>
#include <openssl/hmac.h>
#include <openssl/rand.h>
#else
#define VIIPER_HAS_OPENSSL 0
#endif

namespace viiper
{
    namespace detail
    {

        class EncryptedSocket;
        Result<std::unique_ptr<EncryptedSocket>> perform_handshake(Socket &&socket, const std::string &password);

    } // namespace detail
} // namespace viiper

#include "auth_impl.hpp"
