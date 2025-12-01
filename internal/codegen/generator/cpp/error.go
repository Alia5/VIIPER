package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const errorTemplate = `// Auto-generated VIIPER C++ SDK
// DO NOT EDIT - This file is generated from the VIIPER server codebase

#pragma once

#include <string>
#include <variant>
#include <optional>

namespace viiper {

// ============================================================================
// Error (simplified, generic like Rust SDK)
// ============================================================================

struct Error {
    std::string message;

    Error() = default;
    explicit Error(std::string msg) : message(std::move(msg)) {}

    [[nodiscard]] bool ok() const noexcept { return message.empty(); }
    [[nodiscard]] explicit operator bool() const noexcept { return !ok(); }
    [[nodiscard]] const std::string& to_string() const noexcept { return message; }
};

// ============================================================================
// Result Type (Either value or error)
// ============================================================================

template<typename T>
class Result {
public:
    Result(T value) : data_(std::move(value)) {}
    Result(Error error) : data_(std::move(error)) {}

    [[nodiscard]] bool ok() const noexcept { return std::holds_alternative<T>(data_); }
    [[nodiscard]] bool is_error() const noexcept { return std::holds_alternative<Error>(data_); }
    [[nodiscard]] explicit operator bool() const noexcept { return ok(); }

    [[nodiscard]] T& value() & { return std::get<T>(data_); }
    [[nodiscard]] const T& value() const& { return std::get<T>(data_); }
    [[nodiscard]] T&& value() && { return std::get<T>(std::move(data_)); }

    [[nodiscard]] Error& error() & { return std::get<Error>(data_); }
    [[nodiscard]] const Error& error() const& { return std::get<Error>(data_); }
    [[nodiscard]] Error&& error() && { return std::get<Error>(std::move(data_)); }

    [[nodiscard]] T value_or(T default_value) const& {
        return ok() ? value() : std::move(default_value);
    }

    [[nodiscard]] T* operator->() { return ok() ? &std::get<T>(data_) : nullptr; }
    [[nodiscard]] const T* operator->() const { return ok() ? &std::get<T>(data_) : nullptr; }
    [[nodiscard]] T& operator*() & { return value(); }
    [[nodiscard]] const T& operator*() const& { return value(); }

private:
    std::variant<T, Error> data_;
};

template<>
class Result<void> {
public:
    Result() = default;
    Result(Error error) : error_(std::move(error)) {}

    [[nodiscard]] bool ok() const noexcept { return !error_.has_value(); }
    [[nodiscard]] bool is_error() const noexcept { return error_.has_value(); }
    [[nodiscard]] explicit operator bool() const noexcept { return ok(); }

    [[nodiscard]] Error& error() & { return *error_; }
    [[nodiscard]] const Error& error() const& { return *error_; }

private:
    std::optional<Error> error_;
};

// ============================================================================
// RFC 7807 Problem JSON
// ============================================================================

struct ProblemJson {
    int status = 0;
    std::string title;
    std::string detail;

    [[nodiscard]] bool is_error() const noexcept { return status >= 400; }
    [[nodiscard]] std::string to_string() const {
        return std::to_string(status) + " " + title + (detail.empty() ? "" : ": " + detail);
    }
    [[nodiscard]] Error to_error() const { return Error(to_string()); }
};

} // namespace viiper
`

func generateError(logger *slog.Logger, includeDir string) error {
	logger.Debug("Generating error.hpp")
	outputFile := filepath.Join(includeDir, "error.hpp")

	if err := os.WriteFile(outputFile, []byte(errorTemplate), 0644); err != nil {
		return fmt.Errorf("write error.hpp: %w", err)
	}

	logger.Info("Generated error.hpp", "file", outputFile)
	return nil
}
