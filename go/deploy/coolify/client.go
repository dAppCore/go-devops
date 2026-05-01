package coolify

import (
	"context"
	"sync"

	core "dappco.re/go"
	log "dappco.re/go/log"

	"dappco.re/go/devops/deploy/python"
)

// Client wraps the Python CoolifyClient for Go usage.
type Client struct {
	baseURL   string
	apiToken  string
	timeout   int
	verifySSL bool
	call      func(context.Context, string, map[string]any) (map[string]any, core.Result)

	mu sync.Mutex
}

// Config holds Coolify client configuration.
type Config struct {
	BaseURL   string
	APIToken  string
	Timeout   int
	VerifySSL bool
}

var initEmbeddedPython = python.Init

// DefaultConfig returns default configuration from environment.
func DefaultConfig() Config {
	return Config{
		BaseURL:   core.Getenv("COOLIFY_URL"),
		APIToken:  core.Getenv("COOLIFY_TOKEN"),
		Timeout:   30,
		VerifySSL: true,
	}
}

// NewClient creates a new Coolify client.
func NewClient(cfg Config) (*Client, core.Result) {
	if cfg.BaseURL == "" {
		return nil, core.Fail(log.E("coolify", "COOLIFY_URL not set", nil))
	}
	if cfg.APIToken == "" {
		return nil, core.Fail(log.E("coolify", "COOLIFY_TOKEN not set", nil))
	}

	// Initialize Python runtime
	if r := initEmbeddedPython(); !r.OK {
		return nil, core.Fail(log.E("coolify", "failed to initialize Python", r.Value.(error)))
	}

	return &Client{
		baseURL:   cfg.BaseURL,
		apiToken:  cfg.APIToken,
		timeout:   cfg.Timeout,
		verifySSL: cfg.VerifySSL,
	}, core.Ok(nil)
}

// Call invokes a Coolify API operation by operationId.
func (c *Client) Call(ctx context.Context, operationID string, params map[string]any) (map[string]any, core.Result) {
	if c == nil {
		return nil, core.Fail(log.E("coolify", "client is nil", nil))
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if params == nil {
		params = map[string]any{}
	}
	if c.call != nil {
		return c.call(ctx, operationID, params)
	}

	// Generate and run Python script
	script, r := python.CoolifyScript(c.baseURL, c.apiToken, operationID, params)
	if !r.OK {
		return nil, core.Fail(log.E("coolify", "failed to generate script", r.Value.(error)))
	}
	output, r := python.RunScript(ctx, script)
	if !r.OK {
		return nil, core.Fail(log.E("coolify", "API call "+operationID+" failed", r.Value.(error)))
	}

	// Parse JSON result
	var result map[string]any
	if r := core.JSONUnmarshal([]byte(output), &result); !r.OK {
		// Try parsing as array
		var arrResult []any
		if arr := core.JSONUnmarshal([]byte(output), &arrResult); arr.OK {
			return map[string]any{"result": arrResult}, core.Ok(nil)
		}
		return nil, core.Fail(log.E("coolify", "failed to parse response (output: "+output+")", r.Value.(error)))
	}

	return result, core.Ok(nil)
}

// ListServers returns all servers.
func (c *Client) ListServers(ctx context.Context) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "list-servers", nil)
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// GetServer returns a server by UUID.
func (c *Client) GetServer(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "get-server-by-uuid", map[string]any{"uuid": uuid})
}

// ValidateServer validates a server by UUID.
func (c *Client) ValidateServer(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "validate-server-by-uuid", map[string]any{"uuid": uuid})
}

// ListProjects returns all projects.
func (c *Client) ListProjects(ctx context.Context) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "list-projects", nil)
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// GetProject returns a project by UUID.
func (c *Client) GetProject(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "get-project-by-uuid", map[string]any{"uuid": uuid})
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, name, description string) (map[string]any, core.Result) {
	return c.Call(ctx, "create-project", map[string]any{
		"name":        name,
		"description": description,
	})
}

// ListApplications returns all applications.
func (c *Client) ListApplications(ctx context.Context) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "list-applications", nil)
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// GetApplication returns an application by UUID.
func (c *Client) GetApplication(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "get-application-by-uuid", map[string]any{"uuid": uuid})
}

// DeployApplication triggers deployment of an application.
func (c *Client) DeployApplication(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "deploy-by-tag-or-uuid", map[string]any{"uuid": uuid})
}

// ListDatabases returns all databases.
func (c *Client) ListDatabases(ctx context.Context) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "list-databases", nil)
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// GetDatabase returns a database by UUID.
func (c *Client) GetDatabase(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "get-database-by-uuid", map[string]any{"uuid": uuid})
}

// ListServices returns all services.
func (c *Client) ListServices(ctx context.Context) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "list-services", nil)
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// GetService returns a service by UUID.
func (c *Client) GetService(ctx context.Context, uuid string) (map[string]any, core.Result) {
	return c.Call(ctx, "get-service-by-uuid", map[string]any{"uuid": uuid})
}

// ListEnvironments returns environments for a project.
func (c *Client) ListEnvironments(ctx context.Context, projectUUID string) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "get-environments", map[string]any{"project_uuid": projectUUID})
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// GetTeam returns the current team.
func (c *Client) GetTeam(ctx context.Context) (map[string]any, core.Result) {
	return c.Call(ctx, "get-current-team", nil)
}

// GetTeamMembers returns members of the current team.
func (c *Client) GetTeamMembers(ctx context.Context) ([]map[string]any, core.Result) {
	result, r := c.Call(ctx, "get-current-team-members", nil)
	if !r.OK {
		return nil, r
	}
	return extractArray(result)
}

// extractArray extracts an array from result["result"] or returns empty.
func extractArray(result map[string]any) ([]map[string]any, core.Result) {
	if arr, ok := result["result"].([]any); ok {
		items := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				items = append(items, m)
			}
		}
		return items, core.Ok(nil)
	}
	return nil, core.Ok(nil)
}
