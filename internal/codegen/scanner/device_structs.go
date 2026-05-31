package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ScanDeviceStructs scans a device package for JSON-tagged struct types that can be
// used as typed deviceSpecific helpers in generated SDKs.
func ScanDeviceStructs(devicePkgPath string) ([]DTOSchema, error) {
	entries, err := os.ReadDir(devicePkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", devicePkgPath, err)
	}

	fset := token.NewFileSet()
	var schemas []DTOSchema

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		filePath := filepath.Join(devicePkgPath, entry.Name())
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				schema := DTOSchema{Name: typeSpec.Name.Name}
				hasJSONField := false

				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					fieldName := field.Names[0].Name
					if !ast.IsExported(fieldName) {
						continue
					}

					jsonName, hasJSON, optional := parseJSONTag(field, fieldName)
					if !hasJSON {
						continue
					}
					hasJSONField = true

					typeName, typeKind := extractTypeInfo(field.Type)
					if _, isPtr := field.Type.(*ast.StarExpr); isPtr {
						optional = true
					}

					schema.Fields = append(schema.Fields, FieldInfo{
						Name:     fieldName,
						JSONName: jsonName,
						Type:     typeName,
						TypeKind: typeKind,
						Optional: optional,
					})
				}

				if hasJSONField && len(schema.Fields) > 0 {
					schemas = append(schemas, schema)
				}
			}
		}
	}

	return schemas, nil
}

func parseJSONTag(field *ast.Field, fallback string) (jsonName string, hasJSON bool, optional bool) {
	jsonName = fallback
	if field.Tag == nil {
		return jsonName, false, false
	}

	tag := strings.Trim(field.Tag.Value, "`")
	jsonTag := reflect.StructTag(tag).Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return jsonName, false, false
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] != "" {
		jsonName = parts[0]
	}
	for _, part := range parts[1:] {
		if part == "omitempty" {
			optional = true
			break
		}
	}

	return jsonName, true, optional
}
