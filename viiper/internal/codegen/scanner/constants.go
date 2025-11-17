package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConstantInfo represents a single constant definition
type ConstantInfo struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"` // Can be int, string, etc.
	Type  string      `json:"type"`  // e.g., "int", "uint8", "string"
}

// MapInfo represents a map variable with its entries
type MapInfo struct {
	Name      string                 `json:"name"`
	KeyType   string                 `json:"keyType"`
	ValueType string                 `json:"valueType"`
	Entries   map[string]interface{} `json:"entries"` // Key as string, value as interface{}
}

// DeviceConstants holds all constants and maps for a device package
type DeviceConstants struct {
	DeviceType string         `json:"deviceType"` // "keyboard", "mouse", "xbox360"
	Constants  []ConstantInfo `json:"constants"`
	Maps       []MapInfo      `json:"maps"`
}

// ScanDeviceConstants scans a device package directory for constants and maps
func ScanDeviceConstants(devicePkgPath string) (*DeviceConstants, error) {
	deviceType := filepath.Base(devicePkgPath)
	result := &DeviceConstants{
		DeviceType: deviceType,
		Constants:  []ConstantInfo{},
		Maps:       []MapInfo{},
	}

	entries, err := os.ReadDir(devicePkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", devicePkgPath, err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(devicePkgPath, entry.Name())
		file, err := parseFile(fset, filePath)
		if err != nil {
			continue
		}

		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				if genDecl.Tok == token.CONST {
					result.Constants = append(result.Constants, extractConstants(genDecl)...)
				} else if genDecl.Tok == token.VAR {
					maps := extractMaps(genDecl)
					result.Maps = append(result.Maps, maps...)
				}
			}
		}
	}

	return result, nil
}

func parseFile(fset *token.FileSet, filePath string) (*ast.File, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return parseGoSource(fset, filePath, data)
}

func parseGoSource(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	return parser.ParseFile(fset, filename, src, parser.ParseComments)
}

func extractConstants(genDecl *ast.GenDecl) []ConstantInfo {
	var constants []ConstantInfo

	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		var typeName string
		if valueSpec.Type != nil {
			typeName = exprToString(valueSpec.Type)
		}

		for i, name := range valueSpec.Names {
			if !name.IsExported() {
				continue
			}

			constInfo := ConstantInfo{
				Name: name.Name,
				Type: typeName,
			}

			if i < len(valueSpec.Values) {
				constInfo.Value = extractValue(valueSpec.Values[i])
				if typeName == "" {
					constInfo.Type = inferType(constInfo.Value)
				}
			}

			constants = append(constants, constInfo)
		}
	}

	return constants
}

func extractMaps(genDecl *ast.GenDecl) []MapInfo {
	var maps []MapInfo

	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i, name := range valueSpec.Names {
			if !name.IsExported() {
				continue
			}

			var mapType *ast.MapType

			if valueSpec.Type != nil {
				if mt, ok := valueSpec.Type.(*ast.MapType); ok {
					mapType = mt
				}
			}

			if mapType == nil && i < len(valueSpec.Values) {
				if compositeLit, ok := valueSpec.Values[i].(*ast.CompositeLit); ok {
					if mt, ok := compositeLit.Type.(*ast.MapType); ok {
						mapType = mt
					}
				}
			}

			if mapType == nil {
				continue
			}

			mapInfo := MapInfo{
				Name:      name.Name,
				KeyType:   exprToString(mapType.Key),
				ValueType: exprToString(mapType.Value),
				Entries:   make(map[string]interface{}),
			}

			if i < len(valueSpec.Values) {
				if compositeLit, ok := valueSpec.Values[i].(*ast.CompositeLit); ok {
					mapInfo.Entries = extractMapEntries(compositeLit)
				}
			}

			maps = append(maps, mapInfo)
		}
	}

	return maps
}

func extractMapEntries(compositeLit *ast.CompositeLit) map[string]interface{} {
	entries := make(map[string]interface{})

	for _, elt := range compositeLit.Elts {
		kvExpr, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key := extractValue(kvExpr.Key)
		value := extractValue(kvExpr.Value)

		keyStr := fmt.Sprintf("%v", key)
		entries[keyStr] = value
	}

	return entries
}

func extractValue(expr ast.Expr) interface{} {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			if val, err := strconv.ParseInt(e.Value, 0, 64); err == nil {
				return val
			}
			if val, err := strconv.ParseUint(e.Value, 0, 64); err == nil {
				return val
			}
		case token.STRING:
			return strings.Trim(e.Value, `"`)
		case token.FLOAT:
			if val, err := strconv.ParseFloat(e.Value, 64); err == nil {
				return val
			}
		case token.CHAR:
			val := strings.Trim(e.Value, "'")
			if len(val) == 1 {
				return val
			}
			return val
		}
	case *ast.Ident:
		return e.Name
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s %s %s", extractValue(e.X), e.Op.String(), extractValue(e.Y))
	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%v", e.Op.String(), extractValue(e.X))
	}
	return nil
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		if e.Len != nil {
			return fmt.Sprintf("[%s]%s", exprToString(e.Len), exprToString(e.Elt))
		}
		return "[]" + exprToString(e.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", exprToString(e.Key), exprToString(e.Value))
	}
	return ""
}

func inferType(value interface{}) string {
	switch v := value.(type) {
	case int64:
		if v >= 0 && v <= 255 {
			return "uint8"
		}
		return "int"
	case uint64:
		if v <= 255 {
			return "uint8"
		}
		return "uint"
	case string:
		return "string"
	case float64:
		return "float64"
	default:
		return "unknown"
	}
}
