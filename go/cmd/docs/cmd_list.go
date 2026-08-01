package docs

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
)

// addDocsListCommand adds the 'docs list' command.
func addDocsListCommand(c *core.Core) core.Result {
	return c.Command("docs/list", core.Command{
		Description: i18n.T("cmd.docs.list.short"),
		Flags: core.NewOptions(
			core.Option{Key: "registry", Value: ""},
		),
		Action: func(o core.Options) core.Result {
			return runDocsList(o.String("registry"))
		},
	})
}

func runDocsList(registryPath string) (_ core.Result) {
	reg, _, r := loadRegistry(registryPath)
	if !r.OK {
		return r
	}

	cli.Print("\n%-20s  %-8s  %-8s  %-10s  %s\n",
		headerStyle.Render(i18n.Label("repo")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.readme")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.claude")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.changelog")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.docs")),
	)
	cli.Text("──────────────────────────────────────────────────────────────────────")

	var withDocs, withoutDocs int
	for _, repo := range reg.List() {
		info := scanRepoDocs(repo)

		readme := checkMark(info.Readme != "")
		claude := checkMark(info.ClaudeMd != "")
		changelog := checkMark(info.Changelog != "")

		docsDir := checkMark(false)
		if len(info.DocsFiles) > 0 {
			docsDir = docsFoundStyle.Render(i18n.T("common.count.files", map[string]any{"Count": len(info.DocsFiles)}))
		}

		cli.Print("%-20s  %-8s  %-8s  %-10s  %s\n",
			repoNameStyle.Render(info.Name),
			readme,
			claude,
			changelog,
			docsDir,
		)

		if info.HasDocs {
			withDocs++
		} else {
			withoutDocs++
		}
	}

	cli.Blank()
	cli.Print("%s %s\n",
		cli.KeyStyle.Render(i18n.Label("coverage")),
		i18n.T("cmd.docs.list.coverage_summary", map[string]any{"WithDocs": withDocs, "WithoutDocs": withoutDocs}),
	)

	return core.Ok(nil)
}

func checkMark(ok bool) string {
	if ok {
		return cli.Glyph(":check:")
	}
	return cli.Glyph(":cross:")
}
