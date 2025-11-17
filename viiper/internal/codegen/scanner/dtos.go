package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
)

// DTOSchema represents the schema of a response DTO.
type DTOSchema struct {
	Name   string      `json:"name"`   // Type name (e.g., "BusCreateResponse")
	Fields []FieldInfo `json:"fields"` // Struct fields
}

// FieldInfo describes a single field in a DTO.
type FieldInfo struct {
	Name     string `json:"name"`     // Go field name (e.g., "BusID")
	JSONName string `json:"jsonName"` // JSON tag name (e.g., "busId")
	Type     string `json:"type"`     // Go type (e.g., "uint32", "[]Device")
	TypeKind string `json:"typeKind"` // Kind: "primitive", "slice", "struct"
	Optional bool   `json:"optional"` // Whether field can be omitted (pointer or has omitempty)
}

// ScanDTOs scans a Go file containing DTO struct definitions and extracts their schemas.
func ScanDTOs(filePath string) ([]DTOSchema, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	var schemas []DTOSchema

	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
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

			schema := DTOSchema{
				Name:   typeSpec.Name.Name,
				Fields: []FieldInfo{},
			}

			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}

				fieldName := field.Names[0].Name

				if !ast.IsExported(fieldName) {
					continue
				}

				jsonName := fieldName
				optional := false
				if field.Tag != nil {
					tag := strings.Trim(field.Tag.Value, "`")
					jsonTag := reflect.StructTag(tag).Get("json")
					if jsonTag != "" && jsonTag != "-" {
						parts := strings.Split(jsonTag, ",")
						jsonName = parts[0]
						for _, part := range parts[1:] {
							if part == "omitempty" {
								optional = true
								break
							}
						}
					}
				}

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

			schemas = append(schemas, schema)
		}

		return true
	})

	return schemas, nil
}

func extractTypeInfo(expr ast.Expr) (typeName string, typeKind string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, determineTypeKind(t.Name)
	case *ast.StarExpr:
		innerType, innerKind := extractTypeInfo(t.X)
		return "*" + innerType, innerKind
	case *ast.ArrayType:
		if t.Len == nil {
			elemType, _ := extractTypeInfo(t.Elt)
			return "[]" + elemType, "slice"
		}
		elemType, _ := extractTypeInfo(t.Elt)
		return "[]" + elemType, "array"
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name, "struct"
		}
		return t.Sel.Name, "struct"
	case *ast.MapType:
		keyType, _ := extractTypeInfo(t.Key)
		valueType, _ := extractTypeInfo(t.Value)
		return "map[" + keyType + "]" + valueType, "map"
	default:
		return "unknown", "unknown"
	}
}

func determineTypeKind(typeName string) string {
	primitives := map[string]bool{
		"bool":       true,
		"string":     true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"byte":       true,
		"rune":       true,
		"float32":    true,
		"float64":    true,
		"complex64":  true,
		"complex128": true,
	}

	if primitives[typeName] {
		return "primitive"
	}
	return "struct"
}

// ScanDTOsInPackage scans all Go files in a package and extracts DTO schemas.
func ScanDTOsInPackage(pkgPath string) ([]DTOSchema, error) {
	matches, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("glob package files: %w", err)
	}

	var allSchemas []DTOSchema
	for _, file := range matches {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		schemas, err := ScanDTOs(file)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", file, err)
		}
		allSchemas = append(allSchemas, schemas...)
	}

	return allSchemas, nil
}
