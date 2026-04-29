package coolify

import (
	core "dappco.re/go"
)

func exampleCoolifyClient(result map[string]any) *Client {
	return &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		if result != nil {
			return result, nil
		}
		return map[string]any{"operation": operation, "params": params}, nil
	}}
}

func exampleCoolifyListClient(operation string) *Client {
	return exampleCoolifyClient(map[string]any{"result": []any{map[string]any{"uuid": operation}}})
}

func ExampleDefaultConfig() {
	cfg := DefaultConfig()
	core.Println(cfg.Timeout, cfg.VerifySSL)
	// Output: 30 true
}

func ExampleNewClient() {
	old := initEmbeddedPython
	initEmbeddedPython = func() error { return nil }
	defer func() { initEmbeddedPython = old }()
	client, err := NewClient(Config{BaseURL: "https://coolify.example", APIToken: "token"})
	core.Println(err == nil, client != nil)
	// Output: true true
}

func ExampleClient_Call() {
	client := exampleCoolifyClient(nil)
	result, err := client.Call(core.Background(), "list-servers", nil)
	core.Println(err == nil, result["operation"])
	// Output: true list-servers
}

func ExampleClient_ListServers() {
	items, err := exampleCoolifyListClient("list-servers").ListServers(core.Background())
	core.Println(err == nil, items[0]["uuid"])
	// Output: true list-servers
}

func ExampleClient_GetServer() {
	item, err := exampleCoolifyClient(nil).GetServer(core.Background(), "srv-1")
	core.Println(err == nil, item["operation"], item["params"].(map[string]any)["uuid"])
	// Output: true get-server-by-uuid srv-1
}

func ExampleClient_ValidateServer() {
	item, err := exampleCoolifyClient(nil).ValidateServer(core.Background(), "srv-1")
	core.Println(err == nil, item["operation"])
	// Output: true validate-server-by-uuid
}

func ExampleClient_ListProjects() {
	items, err := exampleCoolifyListClient("list-projects").ListProjects(core.Background())
	core.Println(err == nil, items[0]["uuid"])
	// Output: true list-projects
}

func ExampleClient_GetProject() {
	item, err := exampleCoolifyClient(nil).GetProject(core.Background(), "prj-1")
	core.Println(err == nil, item["operation"])
	// Output: true get-project-by-uuid
}

func ExampleClient_CreateProject() {
	item, err := exampleCoolifyClient(nil).CreateProject(core.Background(), "site", "demo")
	core.Println(err == nil, item["operation"], item["params"].(map[string]any)["name"])
	// Output: true create-project site
}

func ExampleClient_ListApplications() {
	items, err := exampleCoolifyListClient("list-applications").ListApplications(core.Background())
	core.Println(err == nil, items[0]["uuid"])
	// Output: true list-applications
}

func ExampleClient_GetApplication() {
	item, err := exampleCoolifyClient(nil).GetApplication(core.Background(), "app-1")
	core.Println(err == nil, item["operation"])
	// Output: true get-application-by-uuid
}

func ExampleClient_DeployApplication() {
	item, err := exampleCoolifyClient(nil).DeployApplication(core.Background(), "app-1")
	core.Println(err == nil, item["operation"])
	// Output: true deploy-by-tag-or-uuid
}

func ExampleClient_ListDatabases() {
	items, err := exampleCoolifyListClient("list-databases").ListDatabases(core.Background())
	core.Println(err == nil, items[0]["uuid"])
	// Output: true list-databases
}

func ExampleClient_GetDatabase() {
	item, err := exampleCoolifyClient(nil).GetDatabase(core.Background(), "db-1")
	core.Println(err == nil, item["operation"])
	// Output: true get-database-by-uuid
}

func ExampleClient_ListServices() {
	items, err := exampleCoolifyListClient("list-services").ListServices(core.Background())
	core.Println(err == nil, items[0]["uuid"])
	// Output: true list-services
}

func ExampleClient_GetService() {
	item, err := exampleCoolifyClient(nil).GetService(core.Background(), "svc-1")
	core.Println(err == nil, item["operation"])
	// Output: true get-service-by-uuid
}

func ExampleClient_ListEnvironments() {
	items, err := exampleCoolifyListClient("get-environments").ListEnvironments(core.Background(), "prj-1")
	core.Println(err == nil, items[0]["uuid"])
	// Output: true get-environments
}

func ExampleClient_GetTeam() {
	item, err := exampleCoolifyClient(nil).GetTeam(core.Background())
	core.Println(err == nil, item["operation"])
	// Output: true get-current-team
}

func ExampleClient_GetTeamMembers() {
	items, err := exampleCoolifyListClient("get-current-team-members").GetTeamMembers(core.Background())
	core.Println(err == nil, items[0]["uuid"])
	// Output: true get-current-team-members
}
