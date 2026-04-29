// github_labels.go implements GitHub label synchronization.
//
// Uses the gh CLI for label operations:
//   - gh label list --repo {repo} --json name,color,description
//   - gh label create --repo {repo} {name} --color {color} --description {desc}
//   - gh label edit --repo {repo} {name} --color {color} --description {desc}

package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	coreexec "dappco.re/go/process/exec"
)

// GitHubLabel represents a label as returned by the GitHub API.
type GitHubLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// ListLabels fetches all labels for a repository.
func ListLabels(repoFullName string) ([]GitHubLabel, coreFailure) {
	args := []string{
		"label", "list",
		"--repo", repoFullName,
		"--json", "name,color,description",
		"--limit", "200",
	}

	cmd := coreexec.Command(core.Background(), "gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var labels []GitHubLabel
	if r := core.JSONUnmarshal(output, &labels); !r.OK {
		return nil, r.Value.(error)
	}

	return labels, nil
}

// CreateLabel creates a new label in a repository.
func CreateLabel(repoFullName string, label LabelConfig) (_ coreFailure) {
	args := []string{
		"label", "create",
		"--repo", repoFullName,
		label.Name,
		"--color", label.Color,
	}

	if label.Description != "" {
		args = append(args, "--description", label.Description)
	}

	cmd := coreexec.Command(core.Background(), "gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// EditLabel updates an existing label in a repository.
func EditLabel(repoFullName string, label LabelConfig) (_ coreFailure) {
	args := []string{
		"label", "edit",
		"--repo", repoFullName,
		label.Name,
		"--color", label.Color,
	}

	if label.Description != "" {
		args = append(args, "--description", label.Description)
	}

	cmd := coreexec.Command(core.Background(), "gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", core.Trim(string(output)))
	}
	return nil
}

// SyncLabels synchronizes labels for a repository.
// Returns a ChangeSet describing what was changed (or would be changed in dry-run mode).
func SyncLabels(repoFullName string, config *GitHubConfig, dryRun bool) (*ChangeSet, coreFailure) {
	changes := NewChangeSet(repoFullName)

	// Get existing labels
	existing, err := ListLabels(repoFullName)
	if err != nil {
		return nil, cli.Wrap(err, "failed to list labels")
	}

	// Build lookup map
	existingMap := make(map[string]GitHubLabel)
	for _, label := range existing {
		existingMap[core.Lower(label.Name)] = label
	}

	// Process each configured label
	for _, wantLabel := range config.Labels {
		key := core.Lower(wantLabel.Name)
		existing, exists := existingMap[key]

		if !exists {
			// Create new label
			changes.Add(CategoryLabel, ChangeCreate, wantLabel.Name, wantLabel.Description)
			if !dryRun {
				if err := CreateLabel(repoFullName, wantLabel); err != nil {
					return changes, cli.Wrap(err, "failed to create label "+wantLabel.Name)
				}
			}
			continue
		}

		// Check if update is needed
		needsUpdate := false
		details := make(map[string]string)

		if core.Lower(existing.Color) != core.Lower(wantLabel.Color) {
			needsUpdate = true
			details["color"] = existing.Color + " -> " + wantLabel.Color
		}
		if existing.Description != wantLabel.Description {
			needsUpdate = true
			details["description"] = "updated"
		}

		if needsUpdate {
			changes.AddWithDetails(CategoryLabel, ChangeUpdate, wantLabel.Name, "", details)
			if !dryRun {
				if err := EditLabel(repoFullName, wantLabel); err != nil {
					return changes, cli.Wrap(err, "failed to update label "+wantLabel.Name)
				}
			}
		} else {
			changes.Add(CategoryLabel, ChangeSkip, wantLabel.Name, "up to date")
		}
	}

	return changes, nil
}
