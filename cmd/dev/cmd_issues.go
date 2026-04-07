package dev

import (
	"slices"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"

	"dappco.re/go/core/cli/pkg/cli"
	"dappco.re/go/core/i18n"
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

// Issues command flags
var (
	issuesRegistryPath string
	issuesLimit        int
	issuesAssignee     string
)

// addIssuesCommand adds the 'issues' command to the given parent command.
func addIssuesCommand(parent *cli.Command) {
	issuesCmd := &cli.Command{
		Use:   "issues",
		Short: i18n.T("cmd.dev.issues.short"),
		Long:  i18n.T("cmd.dev.issues.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			limit := issuesLimit
			if limit == 0 {
				limit = 10
			}
			return runIssues(issuesRegistryPath, limit, issuesAssignee)
		},
	}

	issuesCmd.Flags().StringVar(&issuesRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	issuesCmd.Flags().IntVarP(&issuesLimit, "limit", "l", 10, i18n.T("cmd.dev.issues.flag.limit"))
	issuesCmd.Flags().StringVarP(&issuesAssignee, "assignee", "a", "", i18n.T("cmd.dev.issues.flag.assignee"))

	parent.AddCommand(issuesCmd)
}

func runIssues(registryPath string, limit int, assignee string) error {
	client, err := forgeAPIClient()
	if err != nil {
		return err
	}

	// Find or use provided registry
	reg, _, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}

	// Fetch issues sequentially
	var allIssues []ForgeIssue
	var fetchErrors []error

	repoList := reg.List()
	for i, repo := range repoList {
		cli.Print("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("i18n.progress.fetch")), i+1, len(repoList), repo.Name)

		owner, apiRepo := forgeRepoIdentity(repo.Path, reg.Org, repo.Name)
		issues, err := fetchIssues(client, owner, apiRepo, repo.Name, limit, assignee)
		if err != nil {
			fetchErrors = append(fetchErrors, cli.Wrap(err, repo.Name))
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
		return nil
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

	return nil
}

func fetchIssues(client *gitea.Client, owner, apiRepo, displayName string, limit int, assignee string) ([]ForgeIssue, error) {
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
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "Not Found") {
			return nil, nil
		}
		return nil, err
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

	return result, nil
}

func printIssue(issue ForgeIssue) {
	// #42 [core-bio] Fix avatar upload
	num := issueNumberStyle.Render(cli.Sprintf("#%d", issue.Number))
	repo := issueRepoStyle.Render(cli.Sprintf("[%s]", issue.RepoName))
	title := issueTitleStyle.Render(cli.Truncate(issue.Title, 60))

	line := cli.Sprintf("  %s %s %s", num, repo, title)

	// Add labels if any
	if len(issue.Labels) > 0 {
		line += " " + issueLabelStyle.Render("["+strings.Join(issue.Labels, ", ")+"]")
	}

	// Add assignee if any
	if len(issue.Assignees) > 0 {
		var tagged []string
		for _, a := range issue.Assignees {
			tagged = append(tagged, "@"+a)
		}
		line += " " + issueAssigneeStyle.Render(strings.Join(tagged, ", "))
	}

	// Add age
	age := cli.FormatAge(issue.CreatedAt)
	line += " " + issueAgeStyle.Render(age)

	cli.Text(line)
}
