package dev

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"dappco.re/go/core/i18n"
	coreio "dappco.re/go/core/io"

	"forge.lthn.ai/core/cli/pkg/cli"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// addSyncCommand adds the 'sync' command to the given parent command.
func addSyncCommand(parent *cli.Command) {
	syncCmd := &cli.Command{
		Use:   "sync",
		Short: i18n.T("cmd.dev.sync.short"),
		Long:  i18n.T("cmd.dev.sync.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			if err := runSync(); err != nil {
				return cli.Wrap(err, i18n.Label("error"))
			}
			cli.Text(i18n.T("i18n.done.sync", "public APIs"))
			return nil
		},
	}

	parent.AddCommand(syncCmd)
}

type symbolInfo struct {
	Name string
	Kind string // "var", "func", "type", "const"
}

func runSync() error {
	pkgDir := "pkg"
	internalDirs, err := coreio.Local.List(pkgDir)
	if err != nil {
		return cli.Wrap(err, "failed to read pkg directory")
	}

	for _, dir := range internalDirs {
		if !dir.IsDir() || dir.Name() == "core" {
			continue
		}

		serviceName := dir.Name()
		internalDir := filepath.Join(pkgDir, serviceName)
		publicDir := serviceName
		publicFile := filepath.Join(publicDir, serviceName+".go")

		if !coreio.Local.Exists(internalDir) {
			continue
		}

		symbols, err := getExportedSymbols(internalDir)
		if err != nil {
			return cli.Wrap(err, cli.Sprintf("error getting symbols for service '%s'", serviceName))
		}

		if err := generatePublicAPIFile(publicDir, publicFile, serviceName, symbols); err != nil {
			return cli.Wrap(err, cli.Sprintf("error generating public API file for service '%s'", serviceName))
		}
	}

	return nil
}

func getExportedSymbols(path string) ([]symbolInfo, error) {
	files, err := listGoFiles(path)
	if err != nil {
		return nil, err
	}

	symbolsByName := make(map[string]symbolInfo)
	for _, file := range files {
		content, err := coreio.Local.Read(file)
		if err != nil {
			return nil, err
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, content, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for name, obj := range node.Scope.Objects {
			if !ast.IsExported(name) {
				continue
			}

			kind := "unknown"
			switch obj.Kind {
			case ast.Con:
				kind = "const"
			case ast.Var:
				kind = "var"
			case ast.Fun:
				kind = "func"
			case ast.Typ:
				kind = "type"
			}

			if kind == "unknown" {
				continue
			}

			if _, exists := symbolsByName[name]; !exists {
				symbolsByName[name] = symbolInfo{Name: name, Kind: kind}
			}
		}
	}

	symbols := make([]symbolInfo, 0, len(symbolsByName))
	for _, symbol := range symbolsByName {
		symbols = append(symbols, symbol)
	}

	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].Name == symbols[j].Name {
			return symbols[i].Kind < symbols[j].Kind
		}
		return symbols[i].Name < symbols[j].Name
	})

	return symbols, nil
}

func listGoFiles(path string) ([]string, error) {
	entries, err := coreio.Local.List(path)
	if err == nil {
		files := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}

			files = append(files, filepath.Join(path, name))
		}
		sort.Strings(files)
		return files, nil
	}

	if coreio.Local.IsFile(path) {
		return []string{path}, nil
	}

	return nil, err
}

const publicAPITemplate = `// package {{.ServiceName}} provides the public API for the {{.ServiceName}} service.
package {{.ServiceName}}

import (
	// Import the internal implementation with an alias.
	impl "forge.lthn.ai/core/cli/{{.ServiceName}}"

	// Import the core contracts to re-export the interface.
	"forge.lthn.ai/core/cli/core"
)

{{range .Symbols}}
{{- if eq .Kind "type"}}
// {{.Name}} is the public type for the {{.Name}} service. It is a type alias
// to the underlying implementation, making it transparent to the user.
type {{.Name}} = impl.{{.Name}}
{{else if eq .Kind "const"}}
// {{.Name}} is a public constant that points to the real constant in the implementation package.
const {{.Name}} = impl.{{.Name}}
{{else if eq .Kind "var"}}
// {{.Name}} is a public variable that points to the real variable in the implementation package.
var {{.Name}} = impl.{{.Name}}
{{else if eq .Kind "func"}}
// {{.Name}} is a public function that points to the real function in the implementation package.
var {{.Name}} = impl.{{.Name}}
{{end}}
{{end}}

// {{.InterfaceName}} is the public interface for the {{.ServiceName}} service.
type {{.InterfaceName}} = core.{{.InterfaceName}}
`

func generatePublicAPIFile(dir, path, serviceName string, symbols []symbolInfo) error {
	if err := coreio.Local.EnsureDir(dir); err != nil {
		return err
	}

	tmpl, err := template.New("publicAPI").Parse(publicAPITemplate)
	if err != nil {
		return err
	}

	tcaser := cases.Title(language.English)
	interfaceName := tcaser.String(serviceName)

	data := struct {
		ServiceName   string
		Symbols       []symbolInfo
		InterfaceName string
	}{
		ServiceName:   serviceName,
		Symbols:       symbols,
		InterfaceName: interfaceName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return coreio.Local.Write(path, buf.String())
}
