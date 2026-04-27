package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	entries, _ := os.ReadDir("lib/viiper")
	fset := token.NewFileSet()
	comments := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		file, _ := parser.ParseFile(fset, "lib/viiper/"+e.Name(), nil, parser.ParseComments)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Doc == nil {
				continue
			}
			var name string
			var lines []string
			for _, c := range fn.Doc.List {
				if n, ok := strings.CutPrefix(c.Text, "//export "); ok {
					name = n
				} else {
					line, ok := strings.CutPrefix(c.Text, "// ")
					if !ok {
						line, _ = strings.CutPrefix(c.Text, "//")
					}
					lines = append(lines, line)
				}
			}
			if name != "" && len(lines) > 0 {
				comments[name] = strings.Join(lines, "\n")
			}
		}
	}

	data, _ := os.ReadFile("dist/libVIIPER/libVIIPER.h")
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "extern ") {
			for _, p := range strings.Fields(line)[1:] {
				if before, _, ok := strings.Cut(p, "("); ok {
					if doc, ok := comments[before]; ok {
						out = append(out, "/*")
						for _, dl := range strings.Split(doc, "\n") {
							out = append(out, " * "+dl)
						}
						out = append(out, " */")
					}
					break
				}
			}
		}
		out = append(out, line)
	}
	os.WriteFile("dist/libVIIPER/libVIIPER.h", []byte(strings.Join(out, "\n")), 0644)
}
