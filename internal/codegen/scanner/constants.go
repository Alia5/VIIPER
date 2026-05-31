package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConstantInfo represents a single constant definition
type ConstantInfo struct {
	Name  string `json:"name"`
	Value any    `json:"value"` // Can be int, string, etc.
	Type  string `json:"type"`  // e.g., "int", "uint8", "string"
}

// MapInfo represents a map variable with its entries
type MapInfo struct {
	Name      string         `json:"name"`
	KeyType   string         `json:"keyType"`
	ValueType string         `json:"valueType"`
	Entries   map[string]any `json:"entries"` // Key as string, value as interface{}
}

// DeviceConstants holds all constants and maps for a device package
type DeviceConstants struct {
	DeviceType string         `json:"deviceType"`
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
	constEnv := make(map[string]ConstantInfo)

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
				switch genDecl.Tok {
				case token.CONST:
					result.Constants = append(result.Constants, extractConstants(genDecl, constEnv)...)
				case token.VAR:
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

func extractConstants(genDecl *ast.GenDecl, env map[string]ConstantInfo) []ConstantInfo {
	var constants []ConstantInfo
	var previousValues []ast.Expr
	previousType := ""
	iotaValue := 0

	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		currentType := previousType
		if len(valueSpec.Values) > 0 {
			currentType = ""
		}
		if valueSpec.Type != nil {
			currentType = exprToString(valueSpec.Type)
		}

		exprs := previousValues
		if len(valueSpec.Values) > 0 {
			exprs = valueSpec.Values
			previousValues = valueSpec.Values
			previousType = currentType
		}

		evaluator := constEvaluator{env: env, iotaValue: iotaValue}

		for i, name := range valueSpec.Names {
			if len(exprs) == 0 {
				continue
			}
			exprIndex := i
			if exprIndex >= len(exprs) {
				exprIndex = len(exprs) - 1
			}

			constInfo := ConstantInfo{
				Name: name.Name,
				Type: currentType,
			}
			constInfo.Value = evaluator.eval(exprs[exprIndex])
			if currentType == "" {
				constInfo.Type = inferType(constInfo.Value)
			}
			env[name.Name] = constInfo

			if name.IsExported() {
				constants = append(constants, constInfo)
			}
		}

		iotaValue++
	}

	return constants
}

type constEvaluator struct {
	env       map[string]ConstantInfo
	iotaValue int
}

func (e constEvaluator) eval(expr ast.Expr) interface{} {
	switch v := expr.(type) {
	case *ast.ParenExpr:
		return e.eval(v.X)
	case *ast.BasicLit:
		switch v.Kind {
		case token.INT:
			if val, err := strconv.ParseInt(v.Value, 0, 64); err == nil {
				return val
			}
			if val, err := strconv.ParseUint(v.Value, 0, 64); err == nil {
				return val
			}
		case token.STRING:
			return strings.Trim(v.Value, `"`)
		case token.FLOAT:
			if val, err := strconv.ParseFloat(v.Value, 64); err == nil {
				return val
			}
		case token.CHAR:
			if unquoted, err := strconv.Unquote(v.Value); err == nil {
				return unquoted
			}
			return strings.Trim(v.Value, "'")
		}
	case *ast.Ident:
		if v.Name == "iota" {
			return int64(e.iotaValue)
		}
		if info, ok := e.env[v.Name]; ok {
			return info.Value
		}
		return v.Name
	case *ast.BinaryExpr:
		left := e.eval(v.X)
		right := e.eval(v.Y)
		if result, ok := evalBinaryExpr(v.Op, left, right); ok {
			return result
		}
		return fmt.Sprintf("%v %s %v", left, v.Op.String(), right)
	case *ast.UnaryExpr:
		value := e.eval(v.X)
		if result, ok := evalUnaryExpr(v.Op, value); ok {
			return result
		}
		return fmt.Sprintf("%s%v", v.Op.String(), value)
	case *ast.SelectorExpr:
		name := exprToString(v)
		if info, ok := e.env[name]; ok {
			return info.Value
		}
		return name
	}
	return nil
}

func evalBinaryExpr(op token.Token, left interface{}, right interface{}) (interface{}, bool) {
	if ls, ok := left.(string); ok && op == token.ADD {
		if rs, ok := right.(string); ok {
			return ls + rs, true
		}
	}

	lu, lok := toUint64(left)
	ru, rok := toUint64(right)
	if lok && rok {
		switch op {
		case token.SHL:
			if ru < 64 {
				return lu << ru, true
			}
		case token.SHR:
			if ru < 64 {
				return lu >> ru, true
			}
		case token.OR:
			return lu | ru, true
		case token.AND:
			return lu & ru, true
		case token.XOR:
			return lu ^ ru, true
		case token.ADD:
			return lu + ru, true
		case token.SUB:
			if lu >= ru {
				return lu - ru, true
			}
		case token.MUL:
			return lu * ru, true
		case token.QUO:
			if ru != 0 {
				return lu / ru, true
			}
		}
	}

	li, lok := toInt64(left)
	ri, rok := toInt64(right)
	if lok && rok {
		switch op {
		case token.ADD:
			return li + ri, true
		case token.SUB:
			return li - ri, true
		case token.MUL:
			return li * ri, true
		case token.QUO:
			if ri != 0 {
				return li / ri, true
			}
		}
	}

	return nil, false
}

func evalUnaryExpr(op token.Token, value interface{}) (interface{}, bool) {
	if u, ok := toUint64(value); ok {
		switch op {
		case token.ADD:
			return u, true
		case token.XOR:
			return ^u, true
		}
	}

	if i, ok := toInt64(value); ok {
		switch op {
		case token.ADD:
			return i, true
		case token.SUB:
			return -i, true
		case token.XOR:
			return ^i, true
		}
	}

	return nil, false
}

func toUint64(value any) (uint64, bool) {
	switch v := value.(type) {
	case uint64:
		return v, true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case string:
		if parsed, err := strconv.ParseUint(v, 0, 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case uint64:
		if v > uint64(math.MaxInt64) {
			return 0, false
		}
		return int64(v), true
	case int:
		return int64(v), true
	case string:
		if parsed, err := strconv.ParseInt(v, 0, 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
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
	return constEvaluator{}.eval(expr)
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
