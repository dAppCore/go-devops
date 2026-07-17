package dev

import (
	"slices"
	"time"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"

	"code.gitea.io/sdk/gitea"
)

// Issue-specific styles (aliases to shared)
var (
	issueRepoStyle     = cli.DimStyle
	issueNumberStyle   = cli.TitleStyle
	issueTitleStyle    = cli.ValueStyle
	issueLabelStyle    = cli.WarningStyle
	issueAssigneeStyle = cli.SuccessStyle
	issueAgeStyle      = cli.DimStyle
)

// ForgeIssue holds display data for an issue.
type ForgeIssue struct {
	Number    int64
	Title     string
	Author    string
	Assignees []string
	Labels    []string
	CreatedAt time.Time
	URL       string
	RepoName  string
}

// addIssuesCommand adds the 'issues' command under "dev".
func addIssuesCommand(c *core.Core) core.Result {
	return c.Command("dev/issues", core.Command{
		Description: i18n.T("cmd.dev.issues.short"),
		Flags: core.NewOptions(
			core.Option{Key: "registry", Value: ""},
			core.Option{Key: "limit", Value: 10},
			core.Option{Key: "assignee", Value: ""},
		),
		Action: func(o core.Options) core.Result {
			limit := o.Int("limit")
			if limit == 0 {
				limit = 10
			}
			return runIssues(o.String("registry"), limit, o.String("assignee"))
		},
	})
}

func runIssues(registryPath string, limit int, assignee string) (_ core.Result) {
	client, r := forgeAPIClient()
	if !r.OK {
		return r
	}

	// Find or use provided registry
	reg, _, r := loadRegistryWithConfig(registryPath)
	if !r.OK {
		return r
	}

	// Fetch issues sequentially
	var allIssues []ForgeIssue
	var fetchErrors []error

	repoList := reg.List()
	for i, repo := range repoList {
		cli.Print("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("i18n.progress.fetch")), i+1, len(repoList), repo.Name)

		owner, apiRepo := forgeRepoIdentity(repo.Path, reg.Org, repo.Name)
		issues, r := fetchIssues(client, owner, apiRepo, repo.Name, limit, assignee)
		if !r.OK {
			fetchErrors = append(fetchErrors, cli.Wrap(r.Value.(error), repo.Name))
			continue
		}
		allIssues = append(allIssues, issues...)
	}
	cli.Print("\033[2K\r") // Clear progress line

	// Sort by created date (newest first)
	slices.SortFunc(allIssues, func(a, b ForgeIssue) int {
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	// Print issues
	if len(allIssues) == 0 {
		cli.Text(i18n.T("cmd.dev.issues.no_issues"))
		return core.Ok(nil)
	}

	cli.Print("\n%s\n\n", i18n.T("cmd.dev.issues.open_issues", map[string]any{"Count": len(allIssues)}))

	for _, issue := range allIssues {
		printIssue(issue)
	}

	// Print any errors
	if len(fetchErrors) > 0 {
		cli.Blank()
		for _, err := range fetchErrors {
			cli.Print("%s %s\n", errorStyle.Render(i18n.Label("error")), err)
		}
	}

	return core.Ok(nil)
}

func fetchIssues(client *gitea.Client, owner, apiRepo, displayName string, limit int, assignee string) ([]ForgeIssue, core.Result) {
	opts := gitea.ListIssueOption{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: limit},
		State:       gitea.StateOpen,
		Type:        gitea.IssueTypeIssue,
	}
	if assignee != "" {
		opts.AssignedBy = assignee
	}

	issues, _, err := client.ListRepoIssues(owner, apiRepo, opts)
	if err != nil {
		errMsg := err.Error()
		if core.Contains(errMsg, "404") || core.Contains(errMsg, "Not Found") {
			return nil, core.Ok(nil)
		}
		return nil, core.Fail(err)
	}

	var result []ForgeIssue
	for _, issue := range issues {
		fi := ForgeIssue{
			Number:    issue.Index,
			Title:     issue.Title,
			CreatedAt: issue.Created,
			URL:       issue.HTMLURL,
			RepoName:  displayName,
		}
		if issue.Poster != nil {
			fi.Author = issue.Poster.UserName
		}
		for _, a := range issue.Assignees {
			fi.Assignees = append(fi.Assignees, a.UserName)
		}
		for _, l := range issue.Labels {
			fi.Labels = append(fi.Labels, l.Name)
		}
		result = append(result, fi)
	}

	return result, core.Ok(nil)
}

func printIssue(issue ForgeIssue) {
	// #42 [core-bio] Fix avatar upload
	num := issueNumberStyle.Render(cli.Sprintf("#%d", issue.Number))
	repo := issueRepoStyle.Render(cli.Sprintf("[%s]", issue.RepoName))
	title := issueTitleStyle.Render(cli.Truncate(issue.Title, 60))

	line := cli.Sprintf("  %s %s %s", num, repo, title)

	// Add labels if any
	if len(issue.Labels) > 0 {
		line += " " + issueLabelStyle.Render("["+core.Join(", ", issue.Labels...)+"]")
	}

	// Add assignee if any
	if len(issue.Assignees) > 0 {
		var tagged []string
		for _, a := range issue.Assignees {
			tagged = append(tagged, "@"+a)
		}
		line += " " + issueAssigneeStyle.Render(core.Join(", ", tagged...))
	}

	// Add age
	age := cli.FormatAge(issue.CreatedAt)
	line += " " + issueAgeStyle.Render(age)

	cli.Text(line)
}
