package dev

import (
	"time"

	"code.gitea.io/sdk/gitea"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
)

// CI-specific styles (aliases to shared)
var (
	ciSuccessStyle = cli.SuccessStyle
	ciFailureStyle = cli.ErrorStyle
	ciPendingStyle = cli.WarningStyle
	ciSkippedStyle = cli.DimStyle
)

// WorkflowRun represents a CI workflow run.
type WorkflowRun struct {
	Name       string
	Status     string
	Conclusion string
	HeadBranch string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	URL        string
	RepoName   string
}

// CI command flags
var (
	ciRegistryPath string
	ciBranch       string
	ciFailedOnly   bool
)

// addCICommand adds the 'ci' command to the given parent command.
func addCICommand(parent *cli.Command) {
	ciCmd := &cli.Command{
		Use:   "ci",
		Short: i18n.T("cmd.dev.ci.short"),
		Long:  i18n.T("cmd.dev.ci.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			branch := ciBranch
			if branch == "" {
				branch = "main"
			}
			return runCI(ciRegistryPath, branch, ciFailedOnly)
		},
	}

	ciCmd.Flags().StringVar(&ciRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	ciCmd.Flags().StringVarP(&ciBranch, "branch", "b", "main", i18n.T("cmd.dev.ci.flag.branch"))
	ciCmd.Flags().BoolVar(&ciFailedOnly, "failed", false, i18n.T("cmd.dev.ci.flag.failed"))

	parent.AddCommand(ciCmd)
}

func runCI(registryPath string, branch string, failedOnly bool) (_ coreFailure) {
	client, err := forgeAPIClient()
	if err != nil {
		return err
	}

	// Find or use provided registry
	reg, _, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}

	// Fetch CI status sequentially
	var allRuns []WorkflowRun
	var fetchErrors []error
	var noCI []string

	repoList := reg.List()
	for i, repo := range repoList {
		cli.Print("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("i18n.progress.check")), i+1, len(repoList), repo.Name)

		owner, apiRepo := forgeRepoIdentity(repo.Path, reg.Org, repo.Name)
		runs, err := fetchWorkflowRuns(client, owner, apiRepo, repo.Name, branch)
		if err != nil {
			if core.Contains(err.Error(), "404") || core.Contains(err.Error(), "no workflows") {
				noCI = append(noCI, repo.Name)
			} else {
				fetchErrors = append(fetchErrors, cli.Wrap(err, repo.Name))
			}
			continue
		}

		if len(runs) > 0 {
			// Just get the latest run
			allRuns = append(allRuns, runs[0])
		} else {
			noCI = append(noCI, repo.Name)
		}
	}
	cli.Print("\033[2K\r") // Clear progress line

	// Count by status
	var success, failed, pending, other int
	for _, run := range allRuns {
		switch run.Conclusion {
		case "success":
			success++
		case "failure":
			failed++
		case "":
			if run.Status == "running" || run.Status == "waiting" ||
				run.Status == "in_progress" || run.Status == "queued" {
				pending++
			} else {
				other++
			}
		default:
			other++
		}
	}

	// Print summary
	cli.Blank()
	cli.Print("%s", i18n.T("cmd.dev.ci.repos_checked", map[string]any{"Count": len(repoList)}))
	if success > 0 {
		cli.Print(" * %s", ciSuccessStyle.Render(i18n.T("cmd.dev.ci.passing", map[string]any{"Count": success})))
	}
	if failed > 0 {
		cli.Print(" * %s", ciFailureStyle.Render(i18n.T("cmd.dev.ci.failing", map[string]any{"Count": failed})))
	}
	if pending > 0 {
		cli.Print(" * %s", ciPendingStyle.Render(i18n.T("common.count.pending", map[string]any{"Count": pending})))
	}
	if len(noCI) > 0 {
		cli.Print(" * %s", ciSkippedStyle.Render(i18n.T("cmd.dev.ci.no_ci", map[string]any{"Count": len(noCI)})))
	}
	cli.Blank()
	cli.Blank()

	// Filter if needed
	displayRuns := allRuns
	if failedOnly {
		displayRuns = nil
		for _, run := range allRuns {
			if run.Conclusion == "failure" {
				displayRuns = append(displayRuns, run)
			}
		}
	}

	// Print details
	for _, run := range displayRuns {
		printWorkflowRun(run)
	}

	// Print errors
	if len(fetchErrors) > 0 {
		cli.Blank()
		for _, err := range fetchErrors {
			cli.Print("%s %s\n", errorStyle.Render(i18n.Label("error")), err)
		}
	}

	return nil
}

func fetchWorkflowRuns(client *gitea.Client, owner, apiRepo, displayName string, branch string) ([]WorkflowRun, coreFailure) {
	// Try ListRepoActionRuns first (Gitea 1.25+ / modern Forgejo)
	resp, _, err := client.ListRepoActionRuns(owner, apiRepo, gitea.ListRepoActionRunsOptions{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: 5},
		Branch:      branch,
	})
	if err == nil && resp != nil && len(resp.WorkflowRuns) > 0 {
		var runs []WorkflowRun
		for _, r := range resp.WorkflowRuns {
			name := r.DisplayTitle
			if r.Path != "" {
				// Use workflow filename as name: ".forgejo/workflows/ci.yml" → "ci"
				name = core.TrimSuffix(core.PathBase(r.Path), core.PathExt(r.Path))
			}
			updated := r.CompletedAt
			if updated.IsZero() {
				updated = r.StartedAt
			}
			runs = append(runs, WorkflowRun{
				Name:       name,
				Status:     r.Status,
				Conclusion: r.Conclusion,
				HeadBranch: r.HeadBranch,
				CreatedAt:  r.StartedAt,
				UpdatedAt:  updated,
				URL:        r.HTMLURL,
				RepoName:   displayName,
			})
		}
		return runs, nil
	}

	// Fallback: ListRepoActionTasks (older API, no version check)
	taskResp, _, taskErr := client.ListRepoActionTasks(owner, apiRepo, gitea.ListOptions{Page: 1, PageSize: 10})
	if taskErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, taskErr
	}

	var runs []WorkflowRun
	for _, t := range taskResp.WorkflowRuns {
		if branch != "" && t.HeadBranch != branch {
			continue
		}
		// ActionTask has single Status field — map to conclusion for completed runs
		conclusion := ""
		switch t.Status {
		case "success", "failure", "skipped", "cancelled":
			conclusion = t.Status
		}
		runs = append(runs, WorkflowRun{
			Name:       t.Name,
			Status:     t.Status,
			Conclusion: conclusion,
			HeadBranch: t.HeadBranch,
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.UpdatedAt,
			URL:        t.URL,
			RepoName:   displayName,
		})
	}

	return runs, nil
}

func printWorkflowRun(run WorkflowRun) {
	// Status icon
	var status string
	switch run.Conclusion {
	case "success":
		status = ciSuccessStyle.Render("v")
	case "failure":
		status = ciFailureStyle.Render("x")
	case "":
		switch run.Status {
		case "running", "in_progress":
			status = ciPendingStyle.Render("*")
		case "waiting", "queued":
			status = ciPendingStyle.Render("o")
		default:
			status = ciSkippedStyle.Render("-")
		}
	case "skipped":
		status = ciSkippedStyle.Render("-")
	case "cancelled":
		status = ciSkippedStyle.Render("o")
	default:
		status = ciSkippedStyle.Render("?")
	}

	// Workflow name (truncated)
	workflowName := cli.Truncate(run.Name, 20)

	// Age
	age := cli.FormatAge(run.UpdatedAt)

	cli.Print("  %s %-18s %-22s %s\n",
		status,
		repoNameStyle.Render(run.RepoName),
		dimStyle.Render(workflowName),
		issueAgeStyle.Render(age),
	)
}
