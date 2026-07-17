module dappco.re/go/devops

go 1.26.2

require (
	dappco.re/go/agent v0.1.0
	dappco.re/go/i18n v0.12.1
	dappco.re/go/io v0.15.1
	dappco.re/go/log v0.13.1
	dappco.re/go/scm v0.22.0
	github.com/kluctl/go-embed-python v0.0.0-3.13.1-20241219-1 // Note: CPython embedding for Ansible playbook execution; no go/* equivalent.
	golang.org/x/term v0.44.0
	golang.org/x/text v0.37.0
	gopkg.in/yaml.v3 v3.0.1 // Note: YAML parser for Ansible inventory and playbook files; no core.* YAML equivalent.
)

require (
	dappco.re/go v0.11.0
	dappco.re/go/cli v0.11.1
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/mattn/go-runewidth v0.0.21 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
)

require dappco.re/go/process v0.16.1

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
)
