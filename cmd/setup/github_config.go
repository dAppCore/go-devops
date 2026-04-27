// github_config.go defines configuration types for GitHub repository setup.
//
// Configuration is loaded from .core/github.yaml and supports environment
// variable expansion using ${VAR} or ${VAR:-default} syntax.

package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	coreio "dappco.re/go/io"
	log "dappco.re/go/log"
	"gopkg.in/yaml.v3"
)

// GitHubConfig represents the full GitHub setup configuration.
type GitHubConfig struct {
	Version          int                      `yaml:"version"`
	Labels           []LabelConfig            `yaml:"labels"`
	Webhooks         map[string]WebhookConfig `yaml:"webhooks"`
	BranchProtection []BranchProtectionConfig `yaml:"branch_protection"`
	Security         SecurityConfig           `yaml:"security"`
}

// LabelConfig defines a GitHub issue/PR label.
type LabelConfig struct {
	Name        string `yaml:"name"`
	Color       string `yaml:"color"`
	Description string `yaml:"description"`
}

// WebhookConfig defines a GitHub webhook configuration.
type WebhookConfig struct {
	URL         string   `yaml:"url"`          // Webhook URL (supports ${ENV_VAR})
	ContentType string   `yaml:"content_type"` // json or form (default: json)
	Secret      string   `yaml:"secret"`       // Optional secret (supports ${ENV_VAR})
	Events      []string `yaml:"events"`       // Events to trigger on
	Active      *bool    `yaml:"active"`       // Whether webhook is active (default: true)
}

// BranchProtectionConfig defines branch protection rules.
type BranchProtectionConfig struct {
	Branch                        string   `yaml:"branch"`
	RequiredReviews               int      `yaml:"required_reviews"`
	DismissStale                  bool     `yaml:"dismiss_stale"`
	RequireCodeOwnerReviews       bool     `yaml:"require_code_owner_reviews"`
	RequiredStatusChecks          []string `yaml:"required_status_checks"`
	RequireLinearHistory          bool     `yaml:"require_linear_history"`
	AllowForcePushes              bool     `yaml:"allow_force_pushes"`
	AllowDeletions                bool     `yaml:"allow_deletions"`
	EnforceAdmins                 bool     `yaml:"enforce_admins"`
	RequireConversationResolution bool     `yaml:"require_conversation_resolution"`
}

// SecurityConfig defines repository security settings.
type SecurityConfig struct {
	DependabotAlerts             bool `yaml:"dependabot_alerts"`
	DependabotSecurityUpdates    bool `yaml:"dependabot_security_updates"`
	SecretScanning               bool `yaml:"secret_scanning"`
	SecretScanningPushProtection bool `yaml:"push_protection"`
}

// LoadGitHubConfig reads and parses a GitHub configuration file.
func LoadGitHubConfig(path string) (*GitHubConfig, error) {
	data, err := coreio.Local.Read(path)
	if err != nil {
		return nil, log.E("setup.github_config", "failed to read config file", err)
	}

	// Expand environment variables before parsing
	expanded := expandEnvVars(data)

	var config GitHubConfig
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, log.E("setup.github_config", "failed to parse config file", err)
	}

	// Set defaults
	for i := range config.Webhooks {
		wh := config.Webhooks[i]
		if wh.ContentType == "" {
			wh.ContentType = "json"
		}
		if wh.Active == nil {
			active := true
			wh.Active = &active
		}
		config.Webhooks[i] = wh
	}

	return &config, nil
}

// envVarPattern matches ${VAR} or ${VAR:-default} patterns.
var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)

// expandEnvVars expands environment variables in the input string.
// Supports ${VAR} and ${VAR:-default} syntax.
func expandEnvVars(input string) string {
	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Parse the match
		submatch := envVarPattern.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		varName := submatch[1]
		defaultValue := ""
		if len(submatch) >= 3 {
			defaultValue = submatch[2]
		}

		// Look up the environment variable
		if value, ok := os.LookupEnv(varName); ok {
			return value
		}
		return defaultValue
	})
}

// FindGitHubConfig searches for github.yaml in common locations.
// Search order:
//  1. Specified path (if non-empty)
//  2. .core/github.yaml (relative to registry)
//  3. github.yaml (relative to registry)
func FindGitHubConfig(registryDir, specifiedPath string) (string, error) {
	if specifiedPath != "" {
		if coreio.Local.IsFile(specifiedPath) {
			return specifiedPath, nil
		}
		return "", log.E("setup.github_config", fmt.Sprintf("config file not found: %s", specifiedPath), nil)
	}

	// Search in common locations (using filepath.Join for OS-portable paths)
	candidates := []string{
		filepath.Join(registryDir, ".core", "github.yaml"),
		filepath.Join(registryDir, "github.yaml"),
	}

	for _, path := range candidates {
		if coreio.Local.IsFile(path) {
			return path, nil
		}
	}

	return "", log.E("setup.github_config", fmt.Sprintf("github.yaml not found in %s/.core/ or %s/", registryDir, registryDir), nil)
}

// Validate checks the configuration for errors.
func (c *GitHubConfig) Validate() error {
	if c.Version != 1 {
		return log.E("setup.github_config", fmt.Sprintf("unsupported config version: %d (expected 1)", c.Version), nil)
	}

	// Validate labels
	for i, label := range c.Labels {
		if label.Name == "" {
			return log.E("setup.github_config", fmt.Sprintf("label %d: name is required", i+1), nil)
		}
		if label.Color == "" {
			return log.E("setup.github_config", fmt.Sprintf("label %q: color is required", label.Name), nil)
		}
		// Validate color format (hex without #)
		if !isValidHexColor(label.Color) {
			return log.E("setup.github_config", fmt.Sprintf("label %q: invalid color %q (expected 6-digit hex without #)", label.Name, label.Color), nil)
		}
	}

	// Validate webhooks (skip those with empty URLs - allows optional webhooks via env vars)
	for name, wh := range c.Webhooks {
		if wh.URL == "" {
			// Empty URL is allowed - webhook will be skipped during sync
			continue
		}
		if len(wh.Events) == 0 {
			return log.E("setup.github_config", fmt.Sprintf("webhook %q: at least one event is required", name), nil)
		}
	}

	// Validate branch protection
	for i, bp := range c.BranchProtection {
		if bp.Branch == "" {
			return log.E("setup.github_config", fmt.Sprintf("branch_protection %d: branch is required", i+1), nil)
		}
	}

	return nil
}

// isValidHexColor checks if a string is a valid 6-digit hex color (without #).
func isValidHexColor(color string) bool {
	if len(color) != 6 {
		return false
	}
	for _, c := range strings.ToLower(color) {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
