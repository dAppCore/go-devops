// Package docs provides documentation management commands.
package docs

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"dappco.re/go/core/i18n"
)

// Style and utility aliases from shared
var (
	repoNameStyle  = cli.RepoStyle
	successStyle   = cli.SuccessStyle
	errorStyle     = cli.ErrorStyle
	dimStyle       = cli.DimStyle
	headerStyle    = cli.HeaderStyle
	confirm        = cli.Confirm
	docsFoundStyle = cli.SuccessStyle
	docsFileStyle  = cli.InfoStyle
)

var docsCmd = &cli.Command{
	Use: "docs",
}

func setDocsI18n() {
	docsCmd.Short = i18n.T("cmd.docs.short")
	docsCmd.Long = i18n.T("cmd.docs.long")
	docsListCmd.Short = i18n.T("cmd.docs.list.short")
	docsListCmd.Long = i18n.T("cmd.docs.list.long")
	docsSyncCmd.Short = i18n.T("cmd.docs.sync.short")
	docsSyncCmd.Long = i18n.T("cmd.docs.sync.long")
}

func init() {
	docsCmd.AddCommand(docsSyncCmd)
	docsCmd.AddCommand(docsListCmd)
}
