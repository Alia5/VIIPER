package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ScanHandlerPayloadInfo analyzes handler functions to infer payload semantics (kind, required, parser hints).
// It complements JSON payload detection; if both numeric and JSON patterns appear JSON wins.
func ScanHandlerPayloadInfo(pkgPath string) (map[string]PayloadInfo, error) {
	matches, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("glob handler files: %w", err)
	}
	out := make(map[string]PayloadInfo)
	for _, file := range matches {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if err := scanPayloadFile(file, out); err != nil {
			return nil, fmt.Errorf("scan %s: %w", file, err)
		}
	}
	return out, nil
}

func scanPayloadFile(filePath string, acc map[string]PayloadInfo) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}

	// Track variable type declarations for JSON target inference
	varDeclTypes := make(map[string]string)
	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || vs.Type == nil {
				continue
			}
			for _, name := range vs.Names {
				varDeclTypes[name.Name] = extractTypeName(vs.Type)
			}
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}
		// Only consider functions returning api.HandlerFunc (SelectorExpr.Sel.Name == HandlerFunc)
		if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
			return true
		}
		returnsHandlerFunc := false
		for _, r := range funcDecl.Type.Results.List {
			if sel, ok := r.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "HandlerFunc" {
				returnsHandlerFunc = true
				break
			}
		}
		if !returnsHandlerFunc {
			return true
		}

		handler := funcDecl.Name.Name
		// Initialize with no payload
		pi := PayloadInfo{Kind: PayloadNone, Required: false}

		// Flags collected
		hasEmptyError := false
		hasNonEmptyBranch := false
		hasJSON := false
		hasNumeric := false
		hasDirectUse := false
		numericBitSize := ""
		jsonTargetType := ""

		// Walk body - also track local variable declarations
		localVarTypes := make(map[string]string)
		ast.Inspect(funcDecl.Body, func(nn ast.Node) bool {
			// Track local variable declarations (var x Type)
			if decl, ok := nn.(*ast.DeclStmt); ok {
				if gen, ok := decl.Decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
					for _, spec := range gen.Specs {
						if vs, ok := spec.(*ast.ValueSpec); ok && vs.Type != nil {
							for _, name := range vs.Names {
								localVarTypes[name.Name] = extractTypeName(vs.Type)
							}
						}
					}
				}
			}

			// If statements for empty/non-empty checks
			if ifs, ok := nn.(*ast.IfStmt); ok {
				if isPayloadComparison(ifs.Cond, token.EQL) || isLenPayloadComparison(ifs.Cond, token.EQL) {
					if blockReturnsError(ifs.Body) {
						hasEmptyError = true
					}
				}
				if isPayloadComparison(ifs.Cond, token.NEQ) || isLenPayloadComparison(ifs.Cond, token.GTR) {
					hasNonEmptyBranch = true
				}
			}

			call, ok := nn.(*ast.CallExpr)
			if ok {
				// Detect json.Unmarshal([]byte(req.Payload), &X)
				if isJSONUnmarshal(call) {
					if len(call.Args) >= 2 {
						if unary, ok := call.Args[1].(*ast.UnaryExpr); ok && unary.Op == token.AND {
							if ident, ok := unary.X.(*ast.Ident); ok {
								// Check local vars first, then package-level vars
								if tname, found := localVarTypes[ident.Name]; found {
									jsonTargetType = baseTypeName(tname)
								} else if tname, found := varDeclTypes[ident.Name]; found {
									jsonTargetType = baseTypeName(tname)
								}
							}
						}
					}
					hasJSON = true
				}
				if isNumericParse(call) {
					hasNumeric = true
					numericBitSize = inferNumericBitSize(call)
				}
				if isFmtSscanf(call) && fmtSscanfUsesPayload(call) {
					hasNumeric = true
					if numericBitSize == "" {
						numericBitSize = "int"
					}
				}
			}

			// Direct usage detection (assignments / passes)
			if assign, ok := nn.(*ast.AssignStmt); ok {
				for _, rhs := range assign.Rhs {
					if usesPayloadDirect(rhs) {
						hasDirectUse = true
					}
				}
			}
			if exprStmt, ok := nn.(*ast.ExprStmt); ok {
				if call, ok := exprStmt.X.(*ast.CallExpr); ok {
					for _, a := range call.Args {
						if usesPayloadDirect(a) {
							hasDirectUse = true
						}
					}
				}
			}
			return true
		})

		// Determine kind precedence: JSON > Numeric > String > None
		switch {
		case hasJSON:
			pi.Kind = PayloadJSON
			pi.Required = hasEmptyError || !hasNonEmptyBranch // current JSON always required
			if jsonTargetType != "" {
				pi.ParserHint, pi.RawType = jsonTargetType, jsonTargetType
			}
			pi.Notes = "JSON payload"
		case hasNumeric:
			pi.Kind = PayloadNumeric
			// Optional if there's a non-empty branch and no empty error
			pi.Required = hasEmptyError || !hasNonEmptyBranch
			if numericBitSize != "" {
				pi.ParserHint, pi.RawType = numericBitSize, numericBitSize
			}
		case hasDirectUse:
			pi.Kind = PayloadString
			pi.Required = hasEmptyError
			pi.ParserHint = "string"
		default:
			// none remains
		}

		acc[handler] = pi
		return true
	})
	return nil
}

// Helper functions
// isPayloadComparison detects req.Payload == "" or != "" depending on op.
func isPayloadComparison(expr ast.Expr, op token.Token) bool {
	be, ok := expr.(*ast.BinaryExpr)
	if !ok || be.Op != op {
		return false
	}
	leftSel, ok := be.X.(*ast.SelectorExpr)
	if !ok || leftSel.Sel.Name != "Payload" {
		return false
	}
	if _, ok := leftSel.X.(*ast.Ident); !ok {
		return false
	}
	lit, ok := be.Y.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING || lit.Value != "\"\"" {
		return false
	}
	return true
}

// isLenPayloadComparison detects len(req.Payload) == 0 or > 0.
func isLenPayloadComparison(expr ast.Expr, op token.Token) bool {
	be, ok := expr.(*ast.BinaryExpr)
	if !ok || be.Op != op {
		return false
	}
	ce, ok := be.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	funIdent, ok := ce.Fun.(*ast.Ident)
	if !ok || funIdent.Name != "len" {
		return false
	}
	if len(ce.Args) != 1 {
		return false
	}
	sel, ok := ce.Args[0].(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Payload" {
		return false
	}
	lit, ok := be.Y.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT {
		return false
	}
	if op == token.EQL && lit.Value == "0" {
		return true
	}
	if op == token.GTR && lit.Value == "0" {
		return true
	} // len(req.Payload) > 0
	return false
}

func blockReturnsError(block *ast.BlockStmt) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.List {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		for _, res := range ret.Results {
			if call, ok := res.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "api" && strings.HasPrefix(sel.Sel.Name, "Err") {
						return true
					}
				}
			}
		}
	}
	return false
}

func isJSONUnmarshal(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if ident, ok := sel.X.(*ast.Ident); !ok || ident.Name != "json" || sel.Sel.Name != "Unmarshal" {
		return false
	}
	if len(call.Args) < 1 {
		return false
	}
	// Expect first arg []byte(req.Payload)
	conv, ok := call.Args[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	if arr, ok := conv.Fun.(*ast.ArrayType); ok {
		if ident, ok := arr.Elt.(*ast.Ident); ok && ident.Name == "byte" {
			if len(conv.Args) == 1 {
				if selExpr, ok := conv.Args[0].(*ast.SelectorExpr); ok && selExpr.Sel.Name == "Payload" {
					return true
				}
			}
		}
	}
	return false
}

func isNumericParse(call *ast.CallExpr) bool {
	funIdent, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := funIdent.X.(*ast.Ident)
	if !ok || ident.Name != "strconv" {
		return false
	}
	switch funIdent.Sel.Name {
	case "ParseUint", "ParseInt", "Atoi":
		// ensure first arg originates from req.Payload (possibly wrapped)
		if len(call.Args) > 0 && originatesFromPayload(call.Args[0]) {
			return true
		}
	}
	return false
}

func inferNumericBitSize(call *ast.CallExpr) string {
	funIdent, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	switch funIdent.Sel.Name {
	case "ParseUint", "ParseInt":
		if len(call.Args) >= 3 { // arg0 string, arg1 base, arg2 bitSize
			if lit, ok := call.Args[2].(*ast.BasicLit); ok && lit.Kind == token.INT {
				return "uint" + lit.Value
			}
		}
		return "int" // fallback
	case "Atoi":
		return "int"
	}
	return ""
}

func originatesFromPayload(expr ast.Expr) bool {
	// Direct req.Payload or wrappers like strings.TrimSpace(req.Payload)
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		if v.Sel.Name == "Payload" {
			return true
		}
	case *ast.CallExpr:
		for _, a := range v.Args {
			if originatesFromPayload(a) {
				return true
			}
		}
	}
	return false
}

func isFmtSscanf(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "fmt" || sel.Sel.Name != "Sscanf" {
		return false
	}
	return true
}

func fmtSscanfUsesPayload(call *ast.CallExpr) bool {
	if len(call.Args) == 0 {
		return false
	}
	return originatesFromPayload(call.Args[0])
}

func usesPayloadDirect(expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		return v.Sel.Name == "Payload"
	case *ast.CallExpr:
		for _, a := range v.Args {
			if usesPayloadDirect(a) {
				return true
			}
		}
	case *ast.UnaryExpr:
		return usesPayloadDirect(v.X)
	case *ast.BinaryExpr:
		return usesPayloadDirect(v.X) || usesPayloadDirect(v.Y)
	}
	return false
}

func extractTypeName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		// e.g., apitypes.DeviceCreateRequest
		left := extractTypeName(v.X)
		if left == "" {
			return v.Sel.Name
		}
		return left + "." + v.Sel.Name
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return extractTypeName(v.X)
	}
	return ""
}

func baseTypeName(full string) string {
	if full == "" {
		return ""
	}
	parts := strings.Split(full, ".")
	return parts[len(parts)-1]
}
