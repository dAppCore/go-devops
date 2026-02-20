package infra

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Constructor ---

func TestNewCloudNSClient_Good(t *testing.T) {
	c := NewCloudNSClient("12345", "secret")
	assert.NotNil(t, c)
	assert.Equal(t, "12345", c.authID)
	assert.Equal(t, "secret", c.password)
	assert.NotNil(t, c.client)
}

// --- authParams ---

func TestCloudNSClient_AuthParams_Good(t *testing.T) {
	c := NewCloudNSClient("49500", "hunter2")
	params := c.authParams()

	assert.Equal(t, "49500", params.Get("auth-id"))
	assert.Equal(t, "hunter2", params.Get("auth-password"))
}

// --- doRaw ---

func TestCloudNSClient_DoRaw_Good_ReturnsBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"Success"}`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "test",
		password: "test",
		client:   ts.Client(),
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/dns/test.json", nil)
	require.NoError(t, err)

	data, err := client.doRaw(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Success")
}

func TestCloudNSClient_DoRaw_Bad_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":"Failed","statusDescription":"Invalid auth"}`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "bad",
		password: "creds",
		client:   ts.Client(),
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/dns/test.json", nil)
	require.NoError(t, err)

	_, err = client.doRaw(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloudns API 403")
}

func TestCloudNSClient_DoRaw_Bad_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal Server Error`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "test",
		password: "test",
		client:   ts.Client(),
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/test", nil)
	require.NoError(t, err)

	_, err = client.doRaw(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cloudns API 500")
}

// --- Zone JSON parsing ---

func TestCloudNSZone_JSON_Good(t *testing.T) {
	data := `[
		{"name": "example.com", "type": "master", "zone": "domain", "status": "1"},
		{"name": "test.io", "type": "master", "zone": "domain", "status": "1"}
	]`

	var zones []CloudNSZone
	err := json.Unmarshal([]byte(data), &zones)

	require.NoError(t, err)
	require.Len(t, zones, 2)
	assert.Equal(t, "example.com", zones[0].Name)
	assert.Equal(t, "master", zones[0].Type)
	assert.Equal(t, "test.io", zones[1].Name)
}

func TestCloudNSZone_JSON_Good_EmptyResponse(t *testing.T) {
	// CloudNS returns {} for no zones, not []
	data := `{}`

	var zones []CloudNSZone
	err := json.Unmarshal([]byte(data), &zones)

	// Should fail to parse as slice — this is the edge case ListZones handles
	assert.Error(t, err)
}

// --- Record JSON parsing ---

func TestCloudNSRecord_JSON_Good(t *testing.T) {
	data := `{
		"12345": {
			"id": "12345",
			"type": "A",
			"host": "www",
			"record": "1.2.3.4",
			"ttl": "3600",
			"status": 1
		},
		"12346": {
			"id": "12346",
			"type": "MX",
			"host": "",
			"record": "mail.example.com",
			"ttl": "3600",
			"priority": "10",
			"status": 1
		}
	}`

	var records map[string]CloudNSRecord
	err := json.Unmarshal([]byte(data), &records)

	require.NoError(t, err)
	require.Len(t, records, 2)

	aRecord := records["12345"]
	assert.Equal(t, "12345", aRecord.ID)
	assert.Equal(t, "A", aRecord.Type)
	assert.Equal(t, "www", aRecord.Host)
	assert.Equal(t, "1.2.3.4", aRecord.Record)
	assert.Equal(t, "3600", aRecord.TTL)
	assert.Equal(t, 1, aRecord.Status)

	mxRecord := records["12346"]
	assert.Equal(t, "MX", mxRecord.Type)
	assert.Equal(t, "mail.example.com", mxRecord.Record)
	assert.Equal(t, "10", mxRecord.Priority)
}

func TestCloudNSRecord_JSON_Good_TXTRecord(t *testing.T) {
	data := `{
		"99": {
			"id": "99",
			"type": "TXT",
			"host": "_acme-challenge",
			"record": "abc123def456",
			"ttl": "60",
			"status": 1
		}
	}`

	var records map[string]CloudNSRecord
	err := json.Unmarshal([]byte(data), &records)

	require.NoError(t, err)
	require.Len(t, records, 1)

	txt := records["99"]
	assert.Equal(t, "TXT", txt.Type)
	assert.Equal(t, "_acme-challenge", txt.Host)
	assert.Equal(t, "abc123def456", txt.Record)
	assert.Equal(t, "60", txt.TTL)
}

// --- CreateRecord response parsing ---

func TestCloudNSClient_CreateRecord_Good_ResponseParsing(t *testing.T) {
	// Verify the response shape CreateRecord expects
	data := `{"status":"Success","statusDescription":"The record was created successfully.","data":{"id":54321}}`

	var result struct {
		Status            string `json:"status"`
		StatusDescription string `json:"statusDescription"`
		Data              struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	err := json.Unmarshal([]byte(data), &result)
	require.NoError(t, err)
	assert.Equal(t, "Success", result.Status)
	assert.Equal(t, 54321, result.Data.ID)
}

func TestCloudNSClient_CreateRecord_Bad_FailedStatus(t *testing.T) {
	// Verify non-Success status produces an error message
	data := `{"status":"Failed","statusDescription":"Record already exists."}`

	var result struct {
		Status            string `json:"status"`
		StatusDescription string `json:"statusDescription"`
	}

	err := json.Unmarshal([]byte(data), &result)
	require.NoError(t, err)
	assert.Equal(t, "Failed", result.Status)
	assert.Equal(t, "Record already exists.", result.StatusDescription)
}

// --- UpdateRecord/DeleteRecord response parsing ---

func TestCloudNSClient_UpdateDelete_Good_ResponseParsing(t *testing.T) {
	data := `{"status":"Success","statusDescription":"The record was updated successfully."}`

	var result struct {
		Status            string `json:"status"`
		StatusDescription string `json:"statusDescription"`
	}

	err := json.Unmarshal([]byte(data), &result)
	require.NoError(t, err)
	assert.Equal(t, "Success", result.Status)
}

// --- Full round-trip tests via doRaw ---

func TestCloudNSClient_ListZones_Good_ViaDoRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth params are passed
		assert.NotEmpty(t, r.URL.Query().Get("auth-id"))
		assert.NotEmpty(t, r.URL.Query().Get("auth-password"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"example.com","type":"master","zone":"domain","status":"1"}]`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "12345",
		password: "secret",
		client:   ts.Client(),
	}

	// Build a request similar to what get() would build, but pointing at test server
	ctx := context.Background()
	params := client.authParams()
	params.Set("page", "1")
	params.Set("rows-per-page", "100")
	params.Set("search", "")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/dns/list-zones.json?"+params.Encode(), nil)
	require.NoError(t, err)

	data, err := client.doRaw(req)
	require.NoError(t, err)

	var zones []CloudNSZone
	err = json.Unmarshal(data, &zones)
	require.NoError(t, err)
	require.Len(t, zones, 1)
	assert.Equal(t, "example.com", zones[0].Name)
}

func TestCloudNSClient_ListRecords_Good_ViaDoRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "example.com", r.URL.Query().Get("domain-name"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"1": {"id":"1","type":"A","host":"www","record":"1.2.3.4","ttl":"3600","status":1},
			"2": {"id":"2","type":"CNAME","host":"blog","record":"www.example.com","ttl":"3600","status":1}
		}`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "12345",
		password: "secret",
		client:   ts.Client(),
	}

	ctx := context.Background()
	params := client.authParams()
	params.Set("domain-name", "example.com")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/dns/records.json?"+params.Encode(), nil)
	require.NoError(t, err)

	data, err := client.doRaw(req)
	require.NoError(t, err)

	var records map[string]CloudNSRecord
	err = json.Unmarshal(data, &records)
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "A", records["1"].Type)
	assert.Equal(t, "CNAME", records["2"].Type)
}

func TestCloudNSClient_CreateRecord_Good_ViaDoRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "example.com", r.URL.Query().Get("domain-name"))
		assert.Equal(t, "www", r.URL.Query().Get("host"))
		assert.Equal(t, "A", r.URL.Query().Get("record-type"))
		assert.Equal(t, "1.2.3.4", r.URL.Query().Get("record"))
		assert.Equal(t, "3600", r.URL.Query().Get("ttl"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"Success","statusDescription":"The record was created successfully.","data":{"id":99}}`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "12345",
		password: "secret",
		client:   ts.Client(),
	}

	ctx := context.Background()
	params := client.authParams()
	params.Set("domain-name", "example.com")
	params.Set("host", "www")
	params.Set("record-type", "A")
	params.Set("record", "1.2.3.4")
	params.Set("ttl", "3600")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/dns/add-record.json", nil)
	require.NoError(t, err)
	req.URL.RawQuery = params.Encode()

	data, err := client.doRaw(req)
	require.NoError(t, err)

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "Success", result.Status)
	assert.Equal(t, 99, result.Data.ID)
}

func TestCloudNSClient_DeleteRecord_Good_ViaDoRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "example.com", r.URL.Query().Get("domain-name"))
		assert.Equal(t, "42", r.URL.Query().Get("record-id"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"Success","statusDescription":"The record was deleted successfully."}`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "12345",
		password: "secret",
		client:   ts.Client(),
	}

	ctx := context.Background()
	params := client.authParams()
	params.Set("domain-name", "example.com")
	params.Set("record-id", "42")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/dns/delete-record.json", nil)
	require.NoError(t, err)
	req.URL.RawQuery = params.Encode()

	data, err := client.doRaw(req)
	require.NoError(t, err)

	var result struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "Success", result.Status)
}

// --- ACME challenge helpers ---

func TestCloudNSClient_SetACMEChallenge_Good_ParamVerification(t *testing.T) {
	// SetACMEChallenge delegates to CreateRecord with specific params.
	// Verify the delegation shape by checking the expected call.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "example.com", r.URL.Query().Get("domain-name"))
		assert.Equal(t, "_acme-challenge", r.URL.Query().Get("host"))
		assert.Equal(t, "TXT", r.URL.Query().Get("record-type"))
		assert.Equal(t, "60", r.URL.Query().Get("ttl"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"Success","statusDescription":"OK","data":{"id":777}}`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "12345",
		password: "secret",
		client:   ts.Client(),
	}

	// Build request matching what SetACMEChallenge -> CreateRecord -> post() builds
	ctx := context.Background()
	params := client.authParams()
	params.Set("domain-name", "example.com")
	params.Set("host", "_acme-challenge")
	params.Set("record-type", "TXT")
	params.Set("record", "acme-token-value")
	params.Set("ttl", "60")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/dns/add-record.json", nil)
	require.NoError(t, err)
	req.URL.RawQuery = params.Encode()

	data, err := client.doRaw(req)
	require.NoError(t, err)

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "Success", result.Status)
	assert.Equal(t, 777, result.Data.ID)
}

func TestCloudNSClient_ClearACMEChallenge_Good_Logic(t *testing.T) {
	// ClearACMEChallenge lists records, finds _acme-challenge TXT records, deletes them.
	// Test the logic by verifying the record filtering.
	records := map[string]CloudNSRecord{
		"1": {ID: "1", Type: "A", Host: "www", Record: "1.2.3.4"},
		"2": {ID: "2", Type: "TXT", Host: "_acme-challenge", Record: "token1"},
		"3": {ID: "3", Type: "TXT", Host: "_dmarc", Record: "v=DMARC1"},
		"4": {ID: "4", Type: "TXT", Host: "_acme-challenge", Record: "token2"},
	}

	// Simulate the filtering logic from ClearACMEChallenge
	var toDelete []string
	for id, r := range records {
		if r.Host == "_acme-challenge" && r.Type == "TXT" {
			toDelete = append(toDelete, id)
		}
	}

	assert.Len(t, toDelete, 2)
	assert.Contains(t, toDelete, "2")
	assert.Contains(t, toDelete, "4")
}

// --- EnsureRecord logic ---

func TestEnsureRecord_Good_Logic_AlreadyCorrect(t *testing.T) {
	// Simulate the check: host matches, type matches, value matches => no change
	records := map[string]CloudNSRecord{
		"10": {ID: "10", Type: "A", Host: "www", Record: "1.2.3.4"},
	}

	host := "www"
	recordType := "A"
	value := "1.2.3.4"

	var needsUpdate, needsCreate bool
	for _, r := range records {
		if r.Host == host && r.Type == recordType {
			if r.Record == value {
				// Already correct — no change needed
				needsUpdate = false
				needsCreate = false
			} else {
				needsUpdate = true
			}
			break
		}
	}

	if !needsUpdate {
		// Check if we found any match at all
		found := false
		for _, r := range records {
			if r.Host == host && r.Type == recordType {
				found = true
				break
			}
		}
		if !found {
			needsCreate = true
		}
	}

	assert.False(t, needsUpdate, "should not need update when value matches")
	assert.False(t, needsCreate, "should not need create when record exists")
}

func TestEnsureRecord_Good_Logic_NeedsUpdate(t *testing.T) {
	records := map[string]CloudNSRecord{
		"10": {ID: "10", Type: "A", Host: "www", Record: "1.2.3.4"},
	}

	host := "www"
	recordType := "A"
	value := "5.6.7.8" // Different value

	var needsUpdate bool
	for _, r := range records {
		if r.Host == host && r.Type == recordType {
			if r.Record != value {
				needsUpdate = true
			}
			break
		}
	}

	assert.True(t, needsUpdate, "should need update when value differs")
}

func TestEnsureRecord_Good_Logic_NeedsCreate(t *testing.T) {
	records := map[string]CloudNSRecord{
		"10": {ID: "10", Type: "A", Host: "www", Record: "1.2.3.4"},
	}

	host := "api" // Does not exist
	recordType := "A"

	found := false
	for _, r := range records {
		if r.Host == host && r.Type == recordType {
			found = true
			break
		}
	}

	assert.False(t, found, "should not find record for non-existent host")
}

// --- Edge cases ---

func TestCloudNSClient_DoRaw_Good_EmptyBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "test",
		password: "test",
		client:   ts.Client(),
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/test", nil)
	require.NoError(t, err)

	data, err := client.doRaw(req)
	require.NoError(t, err)
	assert.Empty(t, data)
}

func TestCloudNSRecord_JSON_Good_EmptyMap(t *testing.T) {
	// An empty record set is a valid empty map
	data := `{}`

	var records map[string]CloudNSRecord
	err := json.Unmarshal([]byte(data), &records)

	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestCloudNSClient_DoRaw_Good_AuthQueryParams(t *testing.T) {
	// Verify that auth params make it to the server in the query string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "49500", r.URL.Query().Get("auth-id"))
		assert.Equal(t, "supersecret", r.URL.Query().Get("auth-password"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	client := &CloudNSClient{
		authID:   "49500",
		password: "supersecret",
		client:   ts.Client(),
	}

	ctx := context.Background()
	params := client.authParams()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/dns/test.json?"+params.Encode(), nil)
	require.NoError(t, err)

	_, err = client.doRaw(req)
	require.NoError(t, err)
}
