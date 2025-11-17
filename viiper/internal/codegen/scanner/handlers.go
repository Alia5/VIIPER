package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

// HandlerArgInfo describes argument usage discovered in a handler function.
type HandlerArgInfo struct {
	HandlerName string          `json:"handlerName"`
	UsesArgs    bool            `json:"usesArgs"`    // Whether req.Args is accessed
	ArgAccesses []ArgAccessInfo `json:"argAccesses"` // Specific arg accesses like req.Args[0]
}

// ArgAccessInfo describes a specific access to req.Args.
type ArgAccessInfo struct {
	Index    int    `json:"index"`    // Index accessed (e.g., 0 for req.Args[0])
	Required bool   `json:"required"` // Whether it's required (checked with len < N)
	VarName  string `json:"varName"`  // Variable name if assigned (e.g., "busId")
}

// ScanHandlerArgs scans handler function implementations to discover argument usage patterns.
// This helps determine what arguments each route expects.
func ScanHandlerArgs(pkgPath string) (map[string]HandlerArgInfo, error) {
	matches, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("glob handler files: %w", err)
	}

	handlerInfo := make(map[string]HandlerArgInfo)

	for _, file := range matches {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		info, err := scanHandlerFile(file)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", file, err)
		}

		for k, v := range info {
			handlerInfo[k] = v
		}
	}

	return handlerInfo, nil
}

func scanHandlerFile(filePath string) (map[string]HandlerArgInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	handlerInfo := make(map[string]HandlerArgInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
			return true
		}

		returnsHandlerFunc := false
		for _, result := range funcDecl.Type.Results.List {
			if selExpr, ok := result.Type.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name == "HandlerFunc" {
					returnsHandlerFunc = true
					break
				}
			}
		}

		if !returnsHandlerFunc {
			return true
		}

		handlerName := funcDecl.Name.Name
		info := HandlerArgInfo{
			HandlerName: handlerName,
			UsesArgs:    false,
			ArgAccesses: []ArgAccessInfo{},
		}

		lengthChecks := findLengthChecks(funcDecl.Body)

		varAssignments := findArgAssignments(funcDecl.Body)

		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			indexExpr, ok := n.(*ast.IndexExpr)
			if !ok {
				return true
			}

			if selExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
				if selExpr.Sel.Name == "Args" {
					info.UsesArgs = true

					if basicLit, ok := indexExpr.Index.(*ast.BasicLit); ok {
						if basicLit.Kind == token.INT {
							idx, _ := strconv.Atoi(basicLit.Value)

							required := isArgRequired(idx, lengthChecks)

							varName := varAssignments[idx]
							if varName == "" {
								varName = fmt.Sprintf("arg%d", idx)
							}

							info.ArgAccesses = append(info.ArgAccesses, ArgAccessInfo{
								Index:    idx,
								Required: required,
								VarName:  varName,
							})
						}
					}
				}
			}

			return true
		})

		if info.UsesArgs || len(info.ArgAccesses) > 0 {
			handlerInfo[handlerName] = info
		}

		return true
	})

	return handlerInfo, nil
}

// findLengthChecks finds `if len(req.Args) < N` or `if len(req.Args) >= N` patterns
// Returns a map of index -> required status
func findLengthChecks(body *ast.BlockStmt) map[int]bool {
	checks := make(map[int]bool)
	if body == nil {
		return checks
	}

	ast.Inspect(body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		binaryExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
		if !ok {
			return true
		}

		var lenValue int
		var isLess bool

		if callExpr, ok := binaryExpr.X.(*ast.CallExpr); ok {
			if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "len" {
				if len(callExpr.Args) > 0 {
					if selExpr, ok := callExpr.Args[0].(*ast.SelectorExpr); ok {
						if selExpr.Sel.Name == "Args" {
							if lit, ok := binaryExpr.Y.(*ast.BasicLit); ok {
								lenValue, _ = strconv.Atoi(lit.Value)
								isLess = (binaryExpr.Op == token.LSS) // <

								if isLess {
									for i := 0; i < lenValue; i++ {
										checks[i] = true
									}
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	return checks
}

func findArgAssignments(body *ast.BlockStmt) map[int]string {
	assignments := make(map[int]string)
	if body == nil {
		return assignments
	}

	ast.Inspect(body, func(n ast.Node) bool {
		assignStmt, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for rhsIdx, rhs := range assignStmt.Rhs {
			var argIdx int = -1
			var found bool

			if indexExpr, ok := rhs.(*ast.IndexExpr); ok {
				if selExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
					if selExpr.Sel.Name == "Args" {
						if lit, ok := indexExpr.Index.(*ast.BasicLit); ok {
							if lit.Kind == token.INT {
								argIdx, _ = strconv.Atoi(lit.Value)
								found = true
							}
						}
					}
				}
			}

			if !found {
				if callExpr, ok := rhs.(*ast.CallExpr); ok {
					for _, arg := range callExpr.Args {
						if indexExpr, ok := arg.(*ast.IndexExpr); ok {
							if selExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
								if selExpr.Sel.Name == "Args" {
									if lit, ok := indexExpr.Index.(*ast.BasicLit); ok {
										if lit.Kind == token.INT {
											argIdx, _ = strconv.Atoi(lit.Value)
											found = true
											break
										}
									}
								}
							}
						}
						if !found {
							if nestedCall, ok := arg.(*ast.CallExpr); ok {
								for _, nestedArg := range nestedCall.Args {
									if indexExpr, ok := nestedArg.(*ast.IndexExpr); ok {
										if selExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
											if selExpr.Sel.Name == "Args" {
												if lit, ok := indexExpr.Index.(*ast.BasicLit); ok {
													if lit.Kind == token.INT {
														argIdx, _ = strconv.Atoi(lit.Value)
														found = true
														break
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			if found && argIdx >= 0 {
				if rhsIdx < len(assignStmt.Lhs) {
					if ident, ok := assignStmt.Lhs[rhsIdx].(*ast.Ident); ok {
						if ident.Name != "err" && ident.Name != "_" {
							assignments[argIdx] = ident.Name
						}
					}
				}
			}
		}

		return true
	})

	return assignments
}

func isArgRequired(idx int, checks map[int]bool) bool {
	return checks[idx]
}
