package cpp

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

const typesTemplate = `{{.Header}}
#pragma once

#include "config.hpp"
#include "detail/json.hpp"
#include <string>
#include <vector>
#include <optional>
#include <cstdint>

namespace viiper {

{{range .DTOs}}
struct {{pascalcase .Name}};
{{end}}

// ============================================================================
// Management API DTOs
// ============================================================================

{{range .DTOs}}
// {{.Name}}
struct {{pascalcase .Name}} {
{{- range .Fields}}
        {{fieldcpptype .}} {{camelcase .Name}};
{{- end}}

    static {{pascalcase .Name}} from_json(const json_type& j) {
        {{pascalcase .Name}} result;
{{- range .Fields}}
{{- if and .Optional (eq .TypeKind "map")}}
        if (j.contains("{{.JSONName}}") && !j["{{.JSONName}}"].is_null()) {
            result.{{camelcase .Name}} = j["{{.JSONName}}"];
        } else {
            result.{{camelcase .Name}} = std::nullopt;
        }
{{- else if .Optional}}
        result.{{camelcase .Name}} = detail::get_optional_field<{{fieldcpptype . | unwrapOptional}}>(j, "{{.JSONName}}");
{{- else if eq .TypeKind "slice"}}
        result.{{camelcase .Name}} = detail::get_array<{{cpptype .Type | sliceElementType}}>(j, "{{.JSONName}}");
{{- else if eq .TypeKind "map"}}
        if (j.contains("{{.JSONName}}")) {
            result.{{camelcase .Name}} = j["{{.JSONName}}"];
        }
{{- else if isCustomType .Type}}
        if (j.contains("{{.JSONName}}")) {
            result.{{camelcase .Name}} = {{cpptype .Type}}::from_json(j["{{.JSONName}}"]);
        }
{{- else}}
        result.{{camelcase .Name}} = j.value("{{.JSONName}}", {{cpptype .Type}}{});
{{- end}}
{{- end}}
        return result;
    }

    [[nodiscard]] json_type to_json() const {
        json_type j;
{{- range .Fields}}
{{- if .Optional}}
        if ({{camelcase .Name}}.has_value()) {
            j["{{.JSONName}}"] = {{camelcase .Name}}.value();
        }
{{- else if eq .TypeKind "slice"}}
        {
            json_type arr = json_type::array();
            for (const auto& item : {{camelcase .Name}}) {
                {{- if isCustomType .Type}}
                arr.push_back(item.to_json());
                {{- else}}
                arr.push_back(item);
                {{- end}}
            }
            j["{{.JSONName}}"] = std::move(arr);
        }
{{- else}}
        j["{{.JSONName}}"] = {{camelcase .Name}};
{{- end}}
{{- end}}
        return j;
    }
};

{{end}}

} // namespace viiper
`

func generateTypes(logger *slog.Logger, includeDir string, md *meta.Metadata) error {
	logger.Debug("Generating types.hpp")
	outputFile := filepath.Join(includeDir, "types.hpp")

	funcs := tplFuncs(md)

	tmpl := template.Must(template.New("types").Funcs(funcs).Parse(typesTemplate))

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create types.hpp: %w", err)
	}
	defer f.Close()

	data := struct {
		Header string
		DTOs   []scanner.DTOSchema
	}{
		Header: writeFileHeader(),
		DTOs:   md.DTOs,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute types template: %w", err)
	}

	logger.Info("Generated types.hpp", "file", outputFile)
	return nil
}
