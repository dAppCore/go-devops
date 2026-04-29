package deploy

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/devops/deploy/coolify"
	"dappco.re/go/i18n"
	log "dappco.re/go/log"
)

var (
	coolifyURL   string
	coolifyToken string
	outputJSON   bool
)

var resultRunE = func(fn func(*cli.Command, []string) core.Result) func(*cli.Command, []string) error {
	return func(cmd *cli.Command, args []string) error {
		r := fn(cmd, args)
		if !r.OK {
			return r.Value.(error)
		}
		return nil
	}
}

// Cmd is the root deploy command.
var Cmd = &cli.Command{
	Use: "deploy",
}

func setDeployI18n() {
	Cmd.Short = i18n.T("cmd.deploy.short")
	Cmd.Long = i18n.T("cmd.deploy.long")
}

var serversCmd = &cli.Command{
	Use:   "servers",
	Short: "List Coolify servers",
	RunE:  resultRunE(runListServers),
}

var projectsCmd = &cli.Command{
	Use:   "projects",
	Short: "List Coolify projects",
	RunE:  resultRunE(runListProjects),
}

var appsCmd = &cli.Command{
	Use:   "apps",
	Short: "List Coolify applications",
	RunE:  resultRunE(runListApps),
}

var dbsCmd = &cli.Command{
	Use:     "databases",
	Short:   "List Coolify databases",
	Aliases: []string{"dbs", "db"},
	RunE:    resultRunE(runListDatabases),
}

var servicesCmd = &cli.Command{
	Use:   "services",
	Short: "List Coolify services",
	RunE:  resultRunE(runListServices),
}

var teamCmd = &cli.Command{
	Use:   "team",
	Short: "Show current team info",
	RunE:  resultRunE(runTeam),
}

var callCmd = &cli.Command{
	Use:   "call <operation> [params-json]",
	Short: "Call any Coolify API operation",
	Args:  cli.RangeArgs(1, 2),
	RunE:  resultRunE(runCall),
}

func init() {
	// Global flags
	Cmd.PersistentFlags().StringVar(&coolifyURL, "url", core.Getenv("COOLIFY_URL"), "Coolify API URL")
	Cmd.PersistentFlags().StringVar(&coolifyToken, "token", core.Getenv("COOLIFY_TOKEN"), "Coolify API token")
	Cmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Add subcommands
	Cmd.AddCommand(serversCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(dbsCmd)
	Cmd.AddCommand(servicesCmd)
	Cmd.AddCommand(teamCmd)
	Cmd.AddCommand(callCmd)
}

func getClient() (*coolify.Client, core.Result) {
	cfg := coolify.Config{
		BaseURL:   coolifyURL,
		APIToken:  coolifyToken,
		Timeout:   30,
		VerifySSL: true,
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = core.Getenv("COOLIFY_URL")
	}
	if cfg.APIToken == "" {
		cfg.APIToken = core.Getenv("COOLIFY_TOKEN")
	}

	return coolify.NewClient(cfg)
}

func outputResult(data any) (_ core.Result) {
	if outputJSON {
		r := core.JSONMarshalIndent(data, "", "  ")
		if !r.OK {
			return r
		}
		if write := core.WriteString(core.Stdout(), string(r.Value.([]byte))+"\n"); !write.OK {
			return write
		}
		return core.Ok(nil)
	}

	// Pretty print based on type
	switch v := data.(type) {
	case []map[string]any:
		for _, item := range v {
			printItem(item)
		}
	case map[string]any:
		printItem(v)
	default:
		cli.Print("%v\n", data)
	}
	return core.Ok(nil)
}

func printItem(item map[string]any) {
	// Common fields to display
	if uuid, ok := item["uuid"].(string); ok {
		cli.Print("%s  ", cli.DimStyle.Render(uuid[:8]))
	}
	if name, ok := item["name"].(string); ok {
		cli.Print("%s", cli.TitleStyle.Render(name))
	}
	if desc, ok := item["description"].(string); ok && desc != "" {
		cli.Print("  %s", cli.DimStyle.Render(desc))
	}
	if status, ok := item["status"].(string); ok {
		switch status {
		case "running":
			cli.Print("  %s", cli.SuccessStyle.Render("●"))
		case "stopped":
			cli.Print("  %s", cli.ErrorStyle.Render("○"))
		default:
			cli.Print("  %s", cli.DimStyle.Render(status))
		}
	}
	core.Println()
}

func runListServers(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return r
	}

	servers, r := client.ListServers(context.Background())
	if !r.OK {
		return r
	}

	if len(servers) == 0 {
		core.Println("No servers found")
		return core.Ok(nil)
	}

	return outputResult(servers)
}

func runListProjects(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return r
	}

	projects, r := client.ListProjects(context.Background())
	if !r.OK {
		return r
	}

	if len(projects) == 0 {
		core.Println("No projects found")
		return core.Ok(nil)
	}

	return outputResult(projects)
}

func runListApps(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return r
	}

	apps, r := client.ListApplications(context.Background())
	if !r.OK {
		return r
	}

	if len(apps) == 0 {
		core.Println("No applications found")
		return core.Ok(nil)
	}

	return outputResult(apps)
}

func runListDatabases(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return r
	}

	dbs, r := client.ListDatabases(context.Background())
	if !r.OK {
		return r
	}

	if len(dbs) == 0 {
		core.Println("No databases found")
		return core.Ok(nil)
	}

	return outputResult(dbs)
}

func runListServices(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return r
	}

	services, r := client.ListServices(context.Background())
	if !r.OK {
		return r
	}

	if len(services) == 0 {
		core.Println("No services found")
		return core.Ok(nil)
	}

	return outputResult(services)
}

func runTeam(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return r
	}

	team, r := client.GetTeam(context.Background())
	if !r.OK {
		return r
	}

	return outputResult(team)
}

func runCall(cmd *cli.Command, args []string) (_ core.Result) {
	client, r := getClient()
	if !r.OK {
		return core.Fail(cli.WrapVerb(r.Value.(error), "initialize", "client"))
	}

	operation := args[0]
	var params map[string]any
	if len(args) > 1 {
		if r := core.JSONUnmarshal([]byte(args[1]), &params); !r.OK {
			return core.Fail(log.E("deploy", "invalid JSON params", r.Value.(error)))
		}
	}

	result, r := client.Call(context.Background(), operation, params)
	if !r.OK {
		return r
	}

	return outputResult(result)
}
