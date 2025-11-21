package scanner

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// RouteInfo describes a discovered API route.
type RouteInfo struct {
	Path        string            `json:"path"`        // e.g., "bus/{id}/list"
	Method      string            `json:"method"`      // "Register" or "RegisterStream"
	Handler     string            `json:"handler"`     // e.g., "BusList"
	PathParams  map[string]string `json:"pathParams"`  // e.g., {"id": "string"}
	ResponseDTO string            `json:"responseDTO"` // Name of DTO type returned (e.g., "BusListResponse"), empty if none
	Payload     PayloadInfo       `json:"payload"`     // payload classification
}

// PayloadKind enumerates recognized payload semantics.
type PayloadKind string

const (
	PayloadNone    PayloadKind = "none"
	PayloadNumeric PayloadKind = "numeric"
	PayloadJSON    PayloadKind = "json"
	PayloadString  PayloadKind = "string"
)

// PayloadInfo describes how a route's req.Payload is interpreted.
type PayloadInfo struct {
	Kind       PayloadKind `json:"kind"`                 // none|numeric|json|string
	Required   bool        `json:"required"`             // true if handler rejects empty payload
	ParserHint string      `json:"parserHint,omitempty"` // e.g., uint32, DeviceCreateRequest, deviceID
	RawType    string      `json:"rawType,omitempty"`    // Underlying Go type name for JSON / numeric width
	Notes      string      `json:"notes,omitempty"`      // Additional guidance for generators
}


// ScanRoutes scans the specified Go file for router.Register() and router.RegisterStream() calls
// and returns metadata about discovered routes.
func ScanRoutes(filePath string) ([]RouteInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	var routes []RouteInfo

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		methodName := selExpr.Sel.Name
		if methodName != "Register" && methodName != "RegisterStream" {
			return true
		}

		if len(callExpr.Args) < 2 {
			return true
		}

		pathLit, ok := callExpr.Args[0].(*ast.BasicLit)
		if !ok || pathLit.Kind != token.STRING {
			return true
		}
		path := strings.Trim(pathLit.Value, `"`)

		handlerName := extractHandlerName(callExpr.Args[1])

		pathParams := extractPathParams(path)

		routes = append(routes, RouteInfo{
			Path:       path,
			Method:     methodName,
			Handler:    handlerName,
			PathParams: pathParams,
		})

		return true
	})

	return routes, nil
}

// extractHandlerName tries to extract the handler function name from a call expression.
// For handler.BusList(usbSrv), it returns "BusList".
func extractHandlerName(expr ast.Expr) string {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return "unknown"
	}

	switch fun := callExpr.Fun.(type) {
	case *ast.SelectorExpr:
		return fun.Sel.Name
	case *ast.Ident:
		return fun.Name
	default:
		return "unknown"
	}
}

// extractPathParams parses a route pattern like "bus/{id}/list" and returns
// a map of parameter names to their types (currently all "string").
func extractPathParams(pattern string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := part[1 : len(part)-1]
			params[paramName] = "string" // API uses string params, converted as needed
		}
	}
	return params
}

// ScanRoutesInPackage scans all Go files in the specified directory (non-recursively)
// and aggregates route information.
func ScanRoutesInPackage(pkgPath string) ([]RouteInfo, error) {
	matches, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("glob package files: %w", err)
	}

	var allRoutes []RouteInfo
	for _, file := range matches {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		routes, err := ScanRoutes(file)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", file, err)
		}
		allRoutes = append(allRoutes, routes...)
	}

	return allRoutes, nil
}

// EnrichRoutesWithHandlerInfo scans handler implementations and enriches routes with argument metadata.
func EnrichRoutesWithHandlerInfo(routes []RouteInfo, handlerPkgPath string) ([]RouteInfo, error) {
	returnTypes, err := ScanHandlerReturnDTOs(handlerPkgPath)
	if err != nil {
		return nil, fmt.Errorf("scan handler return types: %w", err)
	}
	payloadInfo, err := ScanHandlerPayloadInfo(handlerPkgPath)
	if err != nil {
		return nil, fmt.Errorf("scan handler payload info: %w", err)
	}
	enriched := make([]RouteInfo, len(routes))
	for i, route := range routes {
		enriched[i] = route
		if rt, ok := returnTypes[route.Handler]; ok {
			enriched[i].ResponseDTO = rt
		}
		if pi, ok := payloadInfo[route.Handler]; ok {
			enriched[i].Payload = pi
		} else {
			enriched[i].Payload = PayloadInfo{Kind: PayloadNone, Required: false}
		}
	}
	return enriched, nil
}
