package deploy

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/devops/deploy/coolify"
	log "dappco.re/go/log"
)

// getClient builds a Coolify API client from the --url/--token flags,
// falling back to COOLIFY_URL/COOLIFY_TOKEN when the flags are unset.
//
//	client, r := getClient(opts)
//	if !r.OK { return r }
func getClient(opts core.Options) (*coolify.Client, core.Result) {
	cfg := coolify.Config{
		BaseURL:   opts.String("url"),
		APIToken:  opts.String("token"),
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

// outputResult renders data as JSON (--json) or a coloured summary line.
//
//	return outputResult(opts, servers)
func outputResult(opts core.Options, data any) (_ core.Result) {
	if opts.Bool("json") {
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

func runListServers(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
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

	return outputResult(opts, servers)
}

func runListProjects(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
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

	return outputResult(opts, projects)
}

func runListApps(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
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

	return outputResult(opts, apps)
}

func runListDatabases(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
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

	return outputResult(opts, dbs)
}

func runListServices(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
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

	return outputResult(opts, services)
}

func runTeam(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
	if !r.OK {
		return r
	}

	team, r := client.GetTeam(context.Background())
	if !r.OK {
		return r
	}

	return outputResult(opts, team)
}

// runCall dispatches an arbitrary Coolify API operation by name.
// The operation is the command's sole positional argument; the optional
// params-json payload moved to --params because core.Cli.Run only retains
// the last positional under "_arg" (a second bare positional would silently
// clobber the operation name). See migration notes in cmd_commands.go.
func runCall(opts core.Options) (_ core.Result) {
	client, r := getClient(opts)
	if !r.OK {
		return cli.WrapVerb(r.Value.(error), "initialize", "client")
	}

	operation := opts.String("_arg")
	if operation == "" {
		return core.Fail(log.E("deploy.call", "operation is required: core deploy call <operation> [--params=<json>]", nil))
	}

	var params map[string]any
	if raw := opts.String("params"); raw != "" {
		if r := core.JSONUnmarshal([]byte(raw), &params); !r.OK {
			return core.Fail(log.E("deploy", "invalid JSON params", r.Value.(error)))
		}
	}

	result, r := client.Call(context.Background(), operation, params)
	if !r.OK {
		return r
	}

	return outputResult(opts, result)
}
