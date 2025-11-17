package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ScanHandlerReturnDTOs scans handler implementations to find the apitypes.* struct
// passed to json.Marshal, and returns a mapping of handler factory name (e.g., "BusList")
// to DTO type name (e.g., "BusListResponse"). If no DTO is found, the handler is omitted.
func ScanHandlerReturnDTOs(pkgPath string) (map[string]string, error) {
	matches, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("glob handler files: %w", err)
	}

	out := make(map[string]string)

	for _, file := range matches {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse file: %w", err)
		}

		ast.Inspect(node, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
				return true
			}
			returnsHandler := false
			for _, r := range fn.Type.Results.List {
				if sel, ok := r.Type.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "HandlerFunc" {
						returnsHandler = true
						break
					}
				}
			}
			if !returnsHandler {
				return true
			}

			handlerName := fn.Name.Name

			ast.Inspect(fn.Body, func(n2 ast.Node) bool {
				call, ok := n2.(*ast.CallExpr)
				if !ok {
					return true
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "json" && sel.Sel.Name == "Marshal" {
						if len(call.Args) > 0 {
							dto := detectDTOType(call.Args[0])
							if dto != "" {
								out[handlerName] = dto
							}
						}
					}
				}
				return true
			})
			return true
		})
	}

	return out, nil
}

func detectDTOType(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.CompositeLit:
		switch tt := v.Type.(type) {
		case *ast.SelectorExpr:
			if pkg, ok := tt.X.(*ast.Ident); ok && pkg.Name == "apitypes" {
				return tt.Sel.Name
			}
		}
	case *ast.Ident:
		// Heuristic: walk up to its Obj and inspect Decl if available
		if v.Obj != nil && v.Obj.Decl != nil {
			if asn, ok := v.Obj.Decl.(*ast.AssignStmt); ok {
				for _, rhs := range asn.Rhs {
					if lit, ok := rhs.(*ast.CompositeLit); ok {
						if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
							if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "apitypes" {
								return sel.Sel.Name
							}
						}
					}
				}
			}
			if vs, ok := v.Obj.Decl.(*ast.ValueSpec); ok {
				if len(vs.Values) > 0 {
					if lit, ok := vs.Values[0].(*ast.CompositeLit); ok {
						if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
							if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "apitypes" {
								return sel.Sel.Name
							}
						}
					}
				}
				if vs.Type != nil {
					if sel, ok := vs.Type.(*ast.SelectorExpr); ok {
						if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "apitypes" {
							return sel.Sel.Name
						}
					}
				}
			}
		}
	}
	return ""
}
