// Package locales embeds translation files for the go-devops module.
package locales

import "embed"

//go:embed *.json
var FS embed.FS
