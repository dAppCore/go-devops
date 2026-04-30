package dev

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"text/template"

	core "dappco.re/go"
	"dappco.re/go/i18n"
	coreio "dappco.re/go/io"

	"dappco.re/go/cli/pkg/cli"

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
			if r := runSync(); !r.OK {
				return cli.Wrap(r.Value.(error), i18n.Label("error"))
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

func runSync() (_ core.Result) {
	pkgDir := "pkg"
	internalDirs, err := coreio.Local.List(pkgDir)
	if err != nil {
		return core.Fail(cli.Wrap(err, "failed to read pkg directory"))
	}

	for _, dir := range internalDirs {
		if !dir.IsDir() || dir.Name() == "core" {
			continue
		}

		serviceName := dir.Name()
		internalDir := core.PathJoin(pkgDir, serviceName)
		publicDir := serviceName
		publicFile := core.PathJoin(publicDir, serviceName+".go")

		if !coreio.Local.Exists(internalDir) {
			continue
		}

		symbols, r := getExportedSymbols(internalDir)
		if !r.OK {
			return core.Fail(cli.Wrap(r.Value.(error), cli.Sprintf("error getting symbols for service '%s'", serviceName)))
		}

		if r := generatePublicAPIFile(publicDir, publicFile, serviceName, symbols); !r.OK {
			return core.Fail(cli.Wrap(r.Value.(error), cli.Sprintf("error generating public API file for service '%s'", serviceName)))
		}
	}

	return core.Ok(nil)
}

func getExportedSymbols(path string) ([]symbolInfo, core.Result) {
	files, r := listGoFiles(path)
	if !r.OK {
		return nil, r
	}

	symbolsByName := make(map[string]symbolInfo)
	for _, file := range files {
		content, err := coreio.Local.Read(file)
		if err != nil {
			return nil, core.Fail(err)
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, content, parser.ParseComments)
		if err != nil {
			return nil, core.Fail(err)
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

	return symbols, core.Ok(nil)
}

func listGoFiles(path string) ([]string, core.Result) {
	entries, err := coreio.Local.List(path)
	if err == nil {
		files := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !core.HasSuffix(name, ".go") || core.HasSuffix(name, "_test.go") {
				continue
			}

			files = append(files, core.PathJoin(path, name))
		}
		sort.Strings(files)
		return files, core.Ok(nil)
	}

	if coreio.Local.IsFile(path) {
		return []string{path}, core.Ok(nil)
	}

	return nil, core.Fail(err)
}

const publicAPITemplate = `// package {{.ServiceName}} provides the public API for the {{.ServiceName}} service.
package {{.ServiceName}}

import (
	// Import the internal implementation with an alias.
	impl "dappco.re/go/cli/{{.ServiceName}}"

	// Import the core contracts to re-export the interface.
	"dappco.re/go/cli/core"
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

func generatePublicAPIFile(dir, path, serviceName string, symbols []symbolInfo) (_ core.Result) {
	if err := coreio.Local.EnsureDir(dir); err != nil {
		return core.Fail(err)
	}

	tmpl, err := template.New("publicAPI").Parse(publicAPITemplate)
	if err != nil {
		return core.Fail(err)
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

	buf := core.NewBuffer()
	if err := tmpl.Execute(buf, data); err != nil {
		return core.Fail(err)
	}

	return core.ResultOf(nil, coreio.Local.Write(path, buf.String()))
}
