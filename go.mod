module dappco.re/go/devops

go 1.26.2

require (
	code.gitea.io/sdk/gitea v0.24.1 // Note: Gitea SDK for repository and automation API integration; no core.* equivalent.
	dappco.re/go/agent v0.8.0-alpha.1
	dappco.re/go/container v0.8.0-alpha.1
	dappco.re/go/i18n v0.8.0-alpha.1
	dappco.re/go/io v0.8.0-alpha.1
	dappco.re/go/log v0.8.0-alpha.1
	dappco.re/go/scm v0.8.0-alpha.1
	github.com/kluctl/go-embed-python v0.0.0-3.13.1-20241219-1 // Note: CPython embedding for Ansible playbook execution; no go/* equivalent.
	golang.org/x/term v0.42.0
	golang.org/x/text v0.36.0
	gopkg.in/yaml.v3 v3.0.1 // Note: YAML parser for Ansible inventory and playbook files; no core.* YAML equivalent.
)

require (
	codeberg.org/forgejo/go-sdk v0.0.0 // indirect
	dappco.re/go v0.9.0
	dappco.re/go/cli v0.8.0-alpha.1
	dappco.re/go/config v0.8.0-alpha.1 // indirect
	dappco.re/go/inference v0.8.0-alpha.1 // indirect
	github.com/42wim/httpsig v1.2.4 // indirect
	github.com/charmbracelet/x/ansi v0.11.6 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/davidmz/go-pageant v1.0.2 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-fed/httpsig v1.1.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-runewidth v0.0.21 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

require dappco.re/go/process v0.0.0-00010101000000-000000000000

require (
	dappco.re/go/core v0.8.0-alpha.1 // indirect
	dappco.re/go/core/i18n v0.2.3 // indirect
	dappco.re/go/core/inference v0.2.1 // indirect
	dappco.re/go/core/log v0.1.2 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/bubbletea v1.3.10 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
)

replace (
	codeberg.org/forgejo/go-sdk => github.com/dAppCore/go-scm/third_party/forgejo v0.0.0-20260424224729-c5374e1b928e
	dappco.re/go/agent => github.com/dAppCore/agent v0.8.0-alpha.1
	dappco.re/go/ai => github.com/dAppCore/go-ai v0.8.0-alpha.1
	dappco.re/go/api => github.com/dAppCore/api v0.8.0-alpha.1
	dappco.re/go/config => ../go-config
	dappco.re/go/container => ../go-container
	dappco.re/go/forge => github.com/dAppCore/go-forge v0.8.0-alpha.1
	dappco.re/go/i18n => github.com/dAppCore/go-i18n v0.8.0-alpha.1
	dappco.re/go/inference => github.com/dAppCore/go-inference v0.8.0-alpha.1
	dappco.re/go/io => github.com/dAppCore/go-io v0.8.0-alpha.1
	dappco.re/go/log => github.com/dAppCore/go-log v0.8.0-alpha.1
	dappco.re/go/mcp => github.com/dAppCore/mcp v0.8.0-alpha.1
	dappco.re/go/process => ../go-process
	dappco.re/go/rag => github.com/dAppCore/go-rag v0.8.0-alpha.1
	dappco.re/go/scm => github.com/dAppCore/go-scm v0.8.0-alpha.1
	dappco.re/go/store => github.com/dAppCore/go-store v0.8.0-alpha.1
	dappco.re/go/webview => github.com/dAppCore/go-webview v0.8.0-alpha.1
	dappco.re/go/ws => github.com/dAppCore/go-ws v0.8.0-alpha.1
)

replace dappco.re/go => ../go

replace dappco.re/go/cli => dappco.re/go/core/cli v0.5.2
