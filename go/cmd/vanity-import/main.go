// Package main provides a Go vanity import server for dappco.re.
//
// When a Go tool requests ?go-get=1, this server responds with HTML
// containing <meta name="go-import"> tags that map dappco.re module
// paths to their Git repositories on forge.lthn.io.
//
// For browser requests (no ?go-get=1), it redirects to the Forgejo
// repository web UI.
package main

import (
	"net/http"

	core "dappco.re/go"
)

var modules = map[string]string{
	"core":  "host-uk/core",
	"build": "host-uk/build",
}

const (
	forgeBase   = "https://forge.lthn.io"
	vanityHost  = "dappco.re"
	defaultAddr = ":8080"
)

func main() {
	addr := core.Getenv("ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	// Allow overriding forge base URL
	forge := core.Getenv("FORGE_URL")
	if forge == "" {
		forge = forgeBase
	}

	// Parse additional modules from VANITY_MODULES env (format: "mod1=owner/repo,mod2=owner/repo")
	if extra := core.Getenv("VANITY_MODULES"); extra != "" {
		for _, entry := range core.Split(extra, ",") {
			parts := core.SplitN(core.Trim(entry), "=", 2)
			if len(parts) == 2 {
				modules[parts[0]] = parts[1]
			}
		}
	}

	http.HandleFunc("/", handler(forge))

	core.Print(core.Stdout(), "vanity-import listening on %s (%d modules)", addr, len(modules))
	for mod, repo := range modules {
		core.Print(core.Stdout(), "  %s/%s -> %s/%s.git", vanityHost, mod, forge, repo)
	}
	if err := http.ListenAndServe(addr, nil); err != nil {
		core.Print(core.Stderr(), "vanity-import failed: %v", err)
		core.Exit(1)
	}
}

func handler(forge string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the first path segment as the module name
		path := core.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			// Root request — redirect to forge org page
			http.Redirect(w, r, forge+"/host-uk", http.StatusFound)
			return
		}

		// Module is the first path segment (e.g., "core" from "/core/pkg/mcp")
		mod := core.SplitN(path, "/", 2)[0]

		repo, ok := modules[mod]
		if !ok {
			http.NotFound(w, r)
			return
		}

		// If go-get=1, serve the vanity import HTML
		if r.URL.Query().Get("go-get") == "1" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if result := core.WriteString(w, core.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="%s/%s git %s/%s.git">
<meta name="go-source" content="%s/%s %s/%s %s/%s/src/branch/main{/dir} %s/%s/src/branch/main{/dir}/{file}#L{line}">
<meta http-equiv="refresh" content="0; url=%s/%s">
</head>
<body>
Redirecting to <a href="%s/%s">%s/%s</a>...
</body>
</html>
`, vanityHost, mod, forge, repo,
				vanityHost, mod, forge, repo, forge, repo, forge, repo,
				forge, repo,
				forge, repo, forge, repo)); !result.OK {
				return
			}
			return
		}

		// Browser request — redirect to Forgejo
		http.Redirect(w, r, forge+"/"+repo, http.StatusFound)
	}
}
