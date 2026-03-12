package dev

import (
	"slices"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-i18n"
)

// PR-specific styles (aliases to shared)
var (
	prNumberStyle   = cli.NumberStyle
	prTitleStyle    = cli.ValueStyle
	prAuthorStyle   = cli.InfoStyle
	prApprovedStyle = cli.SuccessStyle
	prChangesStyle  = cli.WarningStyle
	prPendingStyle  = cli.DimStyle
	prDraftStyle    = cli.DimStyle
)

// ForgePR holds display data for a pull request.
type ForgePR struct {
	Number         int64
	Title          string
	Draft          bool
	Author         string
	ReviewDecision string // "APPROVED", "CHANGES_REQUESTED", or ""
	CreatedAt      time.Time
	URL            string
	RepoName       string
}

// Reviews command flags
var (
	reviewsRegistryPath string
	reviewsAuthor       string
	reviewsShowAll      bool
)

// addReviewsCommand adds the 'reviews' command to the given parent command.
func addReviewsCommand(parent *cli.Command) {
	reviewsCmd := &cli.Command{
		Use:   "reviews",
		Short: i18n.T("cmd.dev.reviews.short"),
		Long:  i18n.T("cmd.dev.reviews.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runReviews(reviewsRegistryPath, reviewsAuthor, reviewsShowAll)
		},
	}

	reviewsCmd.Flags().StringVar(&reviewsRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	reviewsCmd.Flags().StringVar(&reviewsAuthor, "author", "", i18n.T("cmd.dev.reviews.flag.author"))
	reviewsCmd.Flags().BoolVar(&reviewsShowAll, "all", false, i18n.T("cmd.dev.reviews.flag.all"))

	parent.AddCommand(reviewsCmd)
}

func runReviews(registryPath string, author string, showAll bool) error {
	client, err := forgeAPIClient()
	if err != nil {
		return err
	}

	// Find or use provided registry
	reg, _, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}

	// Fetch PRs sequentially
	var allPRs []ForgePR
	var fetchErrors []error

	repoList := reg.List()
	for i, repo := range repoList {
		cli.Print("\033[2K\r%s %d/%d %s", dimStyle.Render(i18n.T("i18n.progress.fetch")), i+1, len(repoList), repo.Name)

		owner, apiRepo := forgeRepoIdentity(repo.Path, reg.Org, repo.Name)
		prs, err := fetchPRs(client, owner, apiRepo, repo.Name, author)
		if err != nil {
			fetchErrors = append(fetchErrors, cli.Wrap(err, repo.Name))
			continue
		}

		for _, pr := range prs {
			// Filter drafts unless --all
			if !showAll && pr.Draft {
				continue
			}
			allPRs = append(allPRs, pr)
		}
	}
	cli.Print("\033[2K\r") // Clear progress line

	// Sort: pending review first, then by date
	slices.SortFunc(allPRs, func(a, b ForgePR) int {
		// Pending reviews come first
		aPending := a.ReviewDecision == "" || a.ReviewDecision == "REVIEW_REQUIRED"
		bPending := b.ReviewDecision == "" || b.ReviewDecision == "REVIEW_REQUIRED"
		if aPending != bPending {
			if aPending {
				return -1
			}
			return 1
		}
		return b.CreatedAt.Compare(a.CreatedAt)
	})

	// Print PRs
	if len(allPRs) == 0 {
		cli.Text(i18n.T("cmd.dev.reviews.no_prs"))
		return nil
	}

	// Count by status
	var pending, approved, changesRequested int
	for _, pr := range allPRs {
		switch pr.ReviewDecision {
		case "APPROVED":
			approved++
		case "CHANGES_REQUESTED":
			changesRequested++
		default:
			pending++
		}
	}

	cli.Blank()
	cli.Print("%s", i18n.T("cmd.dev.reviews.open_prs", map[string]any{"Count": len(allPRs)}))
	if pending > 0 {
		cli.Print(" * %s", prPendingStyle.Render(i18n.T("common.count.pending", map[string]any{"Count": pending})))
	}
	if approved > 0 {
		cli.Print(" * %s", prApprovedStyle.Render(i18n.T("cmd.dev.reviews.approved", map[string]any{"Count": approved})))
	}
	if changesRequested > 0 {
		cli.Print(" * %s", prChangesStyle.Render(i18n.T("cmd.dev.reviews.changes_requested", map[string]any{"Count": changesRequested})))
	}
	cli.Blank()
	cli.Blank()

	for _, pr := range allPRs {
		printPR(pr)
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

func fetchPRs(client *gitea.Client, owner, apiRepo, displayName string, author string) ([]ForgePR, error) {
	prs, _, err := client.ListRepoPullRequests(owner, apiRepo, gitea.ListPullRequestsOptions{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: 50},
		State:       gitea.StateOpen,
	})
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "Not Found") {
			return nil, nil
		}
		return nil, err
	}

	var result []ForgePR
	for _, pr := range prs {
		// Filter by author if specified
		if author != "" && pr.Poster != nil && pr.Poster.UserName != author {
			continue
		}

		fp := ForgePR{
			Number:   pr.Index,
			Title:    pr.Title,
			Draft:    pr.Draft,
			URL:      pr.HTMLURL,
			RepoName: displayName,
		}
		if pr.Created != nil {
			fp.CreatedAt = *pr.Created
		}
		if pr.Poster != nil {
			fp.Author = pr.Poster.UserName
		}

		// Determine review status
		fp.ReviewDecision = determineReviewDecision(client, owner, apiRepo, pr.Index)

		result = append(result, fp)
	}

	return result, nil
}

// determineReviewDecision fetches reviews for a PR and determines the overall status.
func determineReviewDecision(client *gitea.Client, owner, repo string, prIndex int64) string {
	reviews, _, err := client.ListPullReviews(owner, repo, prIndex, gitea.ListPullReviewsOptions{
		ListOptions: gitea.ListOptions{Page: 1, PageSize: 50},
	})
	if err != nil || len(reviews) == 0 {
		return "" // No reviews = pending
	}

	// Track latest actionable review per reviewer
	latestByReviewer := make(map[string]gitea.ReviewStateType)
	for _, review := range reviews {
		if review.Reviewer == nil {
			continue
		}
		// Only consider approval and change-request reviews (not comments)
		if review.State == gitea.ReviewStateApproved || review.State == gitea.ReviewStateRequestChanges {
			latestByReviewer[review.Reviewer.UserName] = review.State
		}
	}

	if len(latestByReviewer) == 0 {
		return ""
	}

	// If any reviewer requested changes, overall = CHANGES_REQUESTED
	for _, state := range latestByReviewer {
		if state == gitea.ReviewStateRequestChanges {
			return "CHANGES_REQUESTED"
		}
	}

	// All reviewers approved
	return "APPROVED"
}

func printPR(pr ForgePR) {
	// #12 [core-php] Webhook validation
	num := prNumberStyle.Render(cli.Sprintf("#%d", pr.Number))
	repo := issueRepoStyle.Render(cli.Sprintf("[%s]", pr.RepoName))
	title := prTitleStyle.Render(cli.Truncate(pr.Title, 50))
	author := prAuthorStyle.Render("@" + pr.Author)

	// Review status
	var status string
	switch pr.ReviewDecision {
	case "APPROVED":
		status = prApprovedStyle.Render(i18n.T("cmd.dev.reviews.status_approved"))
	case "CHANGES_REQUESTED":
		status = prChangesStyle.Render(i18n.T("cmd.dev.reviews.status_changes"))
	default:
		status = prPendingStyle.Render(i18n.T("cmd.dev.reviews.status_pending"))
	}

	// Draft indicator
	draft := ""
	if pr.Draft {
		draft = prDraftStyle.Render(" " + i18n.T("cmd.dev.reviews.draft"))
	}

	age := cli.FormatAge(pr.CreatedAt)

	cli.Print("  %s %s %s%s %s  %s  %s\n", num, repo, title, draft, author, status, issueAgeStyle.Render(age))
}
