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
	RunE:  runListServers,
}

var projectsCmd = &cli.Command{
	Use:   "projects",
	Short: "List Coolify projects",
	RunE:  runListProjects,
}

var appsCmd = &cli.Command{
	Use:   "apps",
	Short: "List Coolify applications",
	RunE:  runListApps,
}

var dbsCmd = &cli.Command{
	Use:     "databases",
	Short:   "List Coolify databases",
	Aliases: []string{"dbs", "db"},
	RunE:    runListDatabases,
}

var servicesCmd = &cli.Command{
	Use:   "services",
	Short: "List Coolify services",
	RunE:  runListServices,
}

var teamCmd = &cli.Command{
	Use:   "team",
	Short: "Show current team info",
	RunE:  runTeam,
}

var callCmd = &cli.Command{
	Use:   "call <operation> [params-json]",
	Short: "Call any Coolify API operation",
	Args:  cli.RangeArgs(1, 2),
	RunE:  runCall,
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

func getClient() (*coolify.Client, coreFailure) {
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

func outputResult(data any) (_ coreFailure) {
	if outputJSON {
		r := core.JSONMarshalIndent(data, "", "  ")
		if !r.OK {
			return r.Value.(error)
		}
		if write := core.WriteString(core.Stdout(), string(r.Value.([]byte))+"\n"); !write.OK {
			return write.Value.(error)
		}
		return nil
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
	return nil
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

func runListServers(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return err
	}

	servers, err := client.ListServers(context.Background())
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		core.Println("No servers found")
		return nil
	}

	return outputResult(servers)
}

func runListProjects(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return err
	}

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		core.Println("No projects found")
		return nil
	}

	return outputResult(projects)
}

func runListApps(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApplications(context.Background())
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		core.Println("No applications found")
		return nil
	}

	return outputResult(apps)
}

func runListDatabases(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return err
	}

	dbs, err := client.ListDatabases(context.Background())
	if err != nil {
		return err
	}

	if len(dbs) == 0 {
		core.Println("No databases found")
		return nil
	}

	return outputResult(dbs)
}

func runListServices(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return err
	}

	services, err := client.ListServices(context.Background())
	if err != nil {
		return err
	}

	if len(services) == 0 {
		core.Println("No services found")
		return nil
	}

	return outputResult(services)
}

func runTeam(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return err
	}

	team, err := client.GetTeam(context.Background())
	if err != nil {
		return err
	}

	return outputResult(team)
}

func runCall(cmd *cli.Command, args []string) (_ coreFailure) {
	client, err := getClient()
	if err != nil {
		return cli.WrapVerb(err, "initialize", "client")
	}

	operation := args[0]
	var params map[string]any
	if len(args) > 1 {
		if r := core.JSONUnmarshal([]byte(args[1]), &params); !r.OK {
			return log.E("deploy", "invalid JSON params", r.Value.(error))
		}
	}

	result, err := client.Call(context.Background(), operation, params)
	if err != nil {
		return err
	}

	return outputResult(result)
}
