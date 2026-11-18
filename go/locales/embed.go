// Package locales embeds translation files for the go-devops module.
package locales

import (
	"embed"

	"github.com/dappcore/go-i18n/go"
)

//go:embed *.json
var FS embed.FS

func init() {
	i18n.RegisterLocales(FS, ".")
}
