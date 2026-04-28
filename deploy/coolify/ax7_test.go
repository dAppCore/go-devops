package coolify

import core "dappco.re/go"

func ax7StubCoolifyInit(t *core.T, err error) {
	original := initEmbeddedPython
	initEmbeddedPython = func() error { return err }
	t.Cleanup(func() { initEmbeddedPython = original })
}

func TestAX7_DefaultConfig_Good(t *core.T) {
	t.Setenv("COOLIFY_URL", "https://coolify.example")
	t.Setenv("COOLIFY_TOKEN", "secret")
	cfg := DefaultConfig()

	core.AssertEqual(t, "https://coolify.example", cfg.BaseURL)
	core.AssertEqual(t, "secret", cfg.APIToken)
	core.AssertTrue(t, cfg.VerifySSL)
}

func TestAX7_DefaultConfig_Bad(t *core.T) {
	t.Setenv("COOLIFY_URL", "")
	t.Setenv("COOLIFY_TOKEN", "")
	cfg := DefaultConfig()

	core.AssertEqual(t, "", cfg.BaseURL)
	core.AssertEqual(t, "", cfg.APIToken)
	core.AssertEqual(t, 30, cfg.Timeout)
}

func TestAX7_DefaultConfig_Ugly(t *core.T) {
	t.Setenv("COOLIFY_URL", "http://localhost:8000/")
	t.Setenv("COOLIFY_TOKEN", " token with spaces ")
	cfg := DefaultConfig()

	core.AssertEqual(t, "http://localhost:8000/", cfg.BaseURL)
	core.AssertEqual(t, " token with spaces ", cfg.APIToken)
	core.AssertTrue(t, cfg.VerifySSL)
}

func TestAX7_NewClient_Good(t *core.T) {
	ax7StubCoolifyInit(t, nil)
	client, err := NewClient(Config{BaseURL: "https://coolify.example", APIToken: "secret", Timeout: 5})

	core.AssertNoError(t, err)
	core.AssertEqual(t, "https://coolify.example", client.baseURL)
	core.AssertEqual(t, "secret", client.apiToken)
}

func TestAX7_NewClient_Bad(t *core.T) {
	ax7StubCoolifyInit(t, nil)
	client, err := NewClient(Config{APIToken: "secret"})

	core.AssertError(t, err)
	core.AssertNil(t, client)
}

func TestAX7_NewClient_Ugly(t *core.T) {
	ax7StubCoolifyInit(t, core.AnError)
	client, err := NewClient(Config{BaseURL: "https://coolify.example", APIToken: "secret"})

	core.AssertError(t, err)
	core.AssertNil(t, client)
}

func TestAX7_Client_Call_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "list-servers", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"ok": true}, nil
	}}

	result, err := client.Call(core.Background(), "list-servers", nil)
	core.AssertNoError(t, err)
	core.AssertEqual(t, true, result["ok"])
}

func TestAX7_Client_Call_Bad(t *core.T) {
	var client *Client
	result, err := client.Call(core.Background(), "list-servers", nil)

	core.AssertError(t, err)
	core.AssertNil(t, result)
}

func TestAX7_Client_Call_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", operation)
		core.AssertEqual(t, "value", params["key"])
		return map[string]any{}, nil
	}}

	result, err := client.Call(core.Background(), "", map[string]any{"key": "value"})
	core.AssertNoError(t, err)
	core.AssertEmpty(t, result)
}

func TestAX7_Client_ListServers_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "list-servers", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"result": []any{map[string]any{"uuid": "srv-1"}}}, nil
	}}

	items, err := client.ListServers(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "srv-1", items[0]["uuid"])
}

func TestAX7_Client_ListServers_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.ListServers(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_ListServers_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) {
		return map[string]any{"result": "none"}, nil
	}}
	items, err := client.ListServers(core.Background())

	core.AssertNoError(t, err)
	core.AssertNil(t, items)
}

func TestAX7_Client_GetServer_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-server-by-uuid", operation)
		core.AssertEqual(t, "srv-1", params["uuid"])
		return map[string]any{"uuid": "srv-1"}, nil
	}}

	item, err := client.GetServer(core.Background(), "srv-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "srv-1", item["uuid"])
}

func TestAX7_Client_GetServer_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.GetServer(core.Background(), "srv-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_GetServer_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"uuid": ""}, nil
	}}

	item, err := client.GetServer(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["uuid"])
}

func TestAX7_Client_ValidateServer_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "validate-server-by-uuid", operation)
		core.AssertEqual(t, "srv-1", params["uuid"])
		return map[string]any{"valid": true}, nil
	}}

	item, err := client.ValidateServer(core.Background(), "srv-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, true, item["valid"])
}

func TestAX7_Client_ValidateServer_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.ValidateServer(core.Background(), "srv-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_ValidateServer_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"valid": false}, nil
	}}

	item, err := client.ValidateServer(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, false, item["valid"])
}

func TestAX7_Client_ListProjects_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "list-projects", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"result": []any{map[string]any{"uuid": "prj-1"}}}, nil
	}}

	items, err := client.ListProjects(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "prj-1", items[0]["uuid"])
}

func TestAX7_Client_ListProjects_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.ListProjects(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_ListProjects_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return map[string]any{}, nil }}
	items, err := client.ListProjects(core.Background())

	core.AssertNoError(t, err)
	core.AssertNil(t, items)
}

func TestAX7_Client_GetProject_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-project-by-uuid", operation)
		core.AssertEqual(t, "prj-1", params["uuid"])
		return map[string]any{"uuid": "prj-1"}, nil
	}}

	item, err := client.GetProject(core.Background(), "prj-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "prj-1", item["uuid"])
}

func TestAX7_Client_GetProject_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.GetProject(core.Background(), "prj-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_GetProject_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"uuid": ""}, nil
	}}

	item, err := client.GetProject(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["uuid"])
}

func TestAX7_Client_CreateProject_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "create-project", operation)
		core.AssertEqual(t, "agent", params["name"])
		return map[string]any{"name": "agent"}, nil
	}}

	item, err := client.CreateProject(core.Background(), "agent", "desc")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "agent", item["name"])
}

func TestAX7_Client_CreateProject_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.CreateProject(core.Background(), "agent", "desc")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_CreateProject_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["name"])
		core.AssertEqual(t, "", params["description"])
		return map[string]any{"name": ""}, nil
	}}

	item, err := client.CreateProject(core.Background(), "", "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["name"])
}

func TestAX7_Client_ListApplications_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "list-applications", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"result": []any{map[string]any{"uuid": "app-1"}}}, nil
	}}

	items, err := client.ListApplications(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "app-1", items[0]["uuid"])
}

func TestAX7_Client_ListApplications_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.ListApplications(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_ListApplications_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) {
		return map[string]any{"result": []any{"bad"}}, nil
	}}
	items, err := client.ListApplications(core.Background())

	core.AssertNoError(t, err)
	core.AssertEmpty(t, items)
}

func TestAX7_Client_GetApplication_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-application-by-uuid", operation)
		core.AssertEqual(t, "app-1", params["uuid"])
		return map[string]any{"uuid": "app-1"}, nil
	}}

	item, err := client.GetApplication(core.Background(), "app-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "app-1", item["uuid"])
}

func TestAX7_Client_GetApplication_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.GetApplication(core.Background(), "app-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_GetApplication_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"uuid": ""}, nil
	}}

	item, err := client.GetApplication(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["uuid"])
}

func TestAX7_Client_DeployApplication_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "deploy-by-tag-or-uuid", operation)
		core.AssertEqual(t, "app-1", params["uuid"])
		return map[string]any{"deployment": "queued"}, nil
	}}

	item, err := client.DeployApplication(core.Background(), "app-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "queued", item["deployment"])
}

func TestAX7_Client_DeployApplication_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.DeployApplication(core.Background(), "app-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_DeployApplication_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"deployment": ""}, nil
	}}

	item, err := client.DeployApplication(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["deployment"])
}

func TestAX7_Client_ListDatabases_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "list-databases", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"result": []any{map[string]any{"uuid": "db-1"}}}, nil
	}}

	items, err := client.ListDatabases(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "db-1", items[0]["uuid"])
}

func TestAX7_Client_ListDatabases_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.ListDatabases(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_ListDatabases_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) {
		return map[string]any{"result": []any{}}, nil
	}}
	items, err := client.ListDatabases(core.Background())

	core.AssertNoError(t, err)
	core.AssertEmpty(t, items)
}

func TestAX7_Client_GetDatabase_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-database-by-uuid", operation)
		core.AssertEqual(t, "db-1", params["uuid"])
		return map[string]any{"uuid": "db-1"}, nil
	}}

	item, err := client.GetDatabase(core.Background(), "db-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "db-1", item["uuid"])
}

func TestAX7_Client_GetDatabase_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.GetDatabase(core.Background(), "db-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_GetDatabase_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"uuid": ""}, nil
	}}

	item, err := client.GetDatabase(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["uuid"])
}

func TestAX7_Client_ListServices_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "list-services", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"result": []any{map[string]any{"uuid": "svc-1"}}}, nil
	}}

	items, err := client.ListServices(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "svc-1", items[0]["uuid"])
}

func TestAX7_Client_ListServices_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.ListServices(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_ListServices_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) {
		return map[string]any{"result": nil}, nil
	}}
	items, err := client.ListServices(core.Background())

	core.AssertNoError(t, err)
	core.AssertNil(t, items)
}

func TestAX7_Client_GetService_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-service-by-uuid", operation)
		core.AssertEqual(t, "svc-1", params["uuid"])
		return map[string]any{"uuid": "svc-1"}, nil
	}}

	item, err := client.GetService(core.Background(), "svc-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "svc-1", item["uuid"])
}

func TestAX7_Client_GetService_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.GetService(core.Background(), "svc-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_GetService_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["uuid"])
		return map[string]any{"uuid": ""}, nil
	}}

	item, err := client.GetService(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "", item["uuid"])
}

func TestAX7_Client_ListEnvironments_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-environments", operation)
		core.AssertEqual(t, "prj-1", params["project_uuid"])
		return map[string]any{"result": []any{map[string]any{"name": "prod"}}}, nil
	}}

	items, err := client.ListEnvironments(core.Background(), "prj-1")
	core.AssertNoError(t, err)
	core.AssertEqual(t, "prod", items[0]["name"])
}

func TestAX7_Client_ListEnvironments_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.ListEnvironments(core.Background(), "prj-1")

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_ListEnvironments_Ugly(t *core.T) {
	client := &Client{call: func(_ core.Context, _ string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "", params["project_uuid"])
		return map[string]any{"result": []any{}}, nil
	}}

	items, err := client.ListEnvironments(core.Background(), "")
	core.AssertNoError(t, err)
	core.AssertEmpty(t, items)
}

func TestAX7_Client_GetTeam_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-current-team", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"name": "core"}, nil
	}}

	item, err := client.GetTeam(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "core", item["name"])
}

func TestAX7_Client_GetTeam_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	item, err := client.GetTeam(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, item)
}

func TestAX7_Client_GetTeam_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return map[string]any{}, nil }}
	item, err := client.GetTeam(core.Background())

	core.AssertNoError(t, err)
	core.AssertEmpty(t, item)
}

func TestAX7_Client_GetTeamMembers_Good(t *core.T) {
	client := &Client{call: func(_ core.Context, operation string, params map[string]any) (map[string]any, error) {
		core.AssertEqual(t, "get-current-team-members", operation)
		core.AssertEmpty(t, params)
		return map[string]any{"result": []any{map[string]any{"name": "alice"}}}, nil
	}}

	items, err := client.GetTeamMembers(core.Background())
	core.AssertNoError(t, err)
	core.AssertEqual(t, "alice", items[0]["name"])
}

func TestAX7_Client_GetTeamMembers_Bad(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) { return nil, core.AnError }}
	items, err := client.GetTeamMembers(core.Background())

	core.AssertErrorIs(t, err, core.AnError)
	core.AssertNil(t, items)
}

func TestAX7_Client_GetTeamMembers_Ugly(t *core.T) {
	client := &Client{call: func(core.Context, string, map[string]any) (map[string]any, error) {
		return map[string]any{"result": []any{}}, nil
	}}
	items, err := client.GetTeamMembers(core.Background())

	core.AssertNoError(t, err)
	core.AssertEmpty(t, items)
}
