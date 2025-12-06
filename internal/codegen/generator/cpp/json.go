package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const jsonTemplate = `// Auto-generated VIIPER C++ Client Library
// DO NOT EDIT - This file is generated from the VIIPER server codebase

#pragma once

#include "../config.hpp"
#include "../error.hpp"
#include <string>
#include <optional>
#include <vector>
#include <type_traits>

namespace viiper {
namespace detail {

// ============================================================================
// JSON parsing helpers
// ============================================================================

inline ProblemJson parse_problem_json(const json_type& j) {
    ProblemJson problem;
    if (j.contains("status") && j["status"].is_number()) {
        problem.status = j["status"].template get<int>();
    }
    if (j.contains("title") && j["title"].is_string()) {
        problem.title = j["title"].template get<std::string>();
    }
    if (j.contains("detail") && j["detail"].is_string()) {
        problem.detail = j["detail"].template get<std::string>();
    }
    return problem;
}

inline bool is_problem_json(const json_type& j) {
    return j.is_object() && j.contains("status") && j["status"].is_number() &&
           j.contains("title") && j["title"].is_string();
}

inline Result<json_type> parse_json_response(const std::string& response) {
    if (response.empty()) {
        return Error("empty response");
    }

    try {
        auto j = json_type::parse(response);
        if (is_problem_json(j)) {
            auto problem = parse_problem_json(j);
            if (problem.is_error()) {
                return problem.to_error();
            }
        }
        return j;
    } catch (const std::exception& e) {
        return Error(std::string("parse error: ") + e.what());
    }
}

template<typename T>
inline std::optional<T> get_optional_field(const json_type& j, const std::string& key) {
    if (j.contains(key) && !j[key].is_null()) {
        return j[key].template get<T>();
    }
    return std::nullopt;
}

template<typename T, typename = void>
struct has_from_json : std::false_type {};

template<typename T>
struct has_from_json<T, std::void_t<decltype(T::from_json(std::declval<const json_type&>()))>> : std::true_type {};

template<typename T>
inline std::vector<T> get_array(const json_type& j, const std::string& key) {
    if (!j.contains(key) || !j[key].is_array()) {
        return {};
    }

    std::vector<T> result;
    const auto& arr = j[key];
    result.reserve(arr.size());

    for (const auto& item : arr) {
        if constexpr (has_from_json<T>::value) {
            result.push_back(T::from_json(item));
        } else {
            result.push_back(item.template get<T>());
        }
    }

    return result;
}

} // namespace detail
} // namespace viiper
`

func generateJson(logger *slog.Logger, detailDir string) error {
	logger.Debug("Generating detail/json.hpp")
	outputFile := filepath.Join(detailDir, "json.hpp")

	if err := os.WriteFile(outputFile, []byte(jsonTemplate), 0644); err != nil {
		return fmt.Errorf("write json.hpp: %w", err)
	}

	logger.Info("Generated json.hpp", "file", outputFile)
	return nil
}
