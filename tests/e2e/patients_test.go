package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/httpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type tenantTransport struct {
	base     http.RoundTripper
	tenantID string
}

func (t tenantTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()
	cloned.Header.Set("X-Tenant-ID", t.tenantID)
	return t.base.RoundTrip(cloned)
}

func connectTestClient(
	t *testing.T,
	endpoint string,
	tenantID string,
) *mcp.ClientSession {
	t.Helper()

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "healthcare-e2e-test",
			Version: "1.0.0",
		},
		nil,
	)

	httpClient := &http.Client{
		Transport: tenantTransport{
			base:     http.DefaultTransport,
			tenantID: tenantID,
		},
		Timeout: 5 * time.Second,
	}

	transport := &mcp.StreamableClientTransport{
		Endpoint:   endpoint,
		HTTPClient: httpClient,
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}

	t.Cleanup(func() {
		_ = session.Close()
	})

	return session
}

func TestSearchPatientsThroughMCP(t *testing.T) {
	router := httpserver.NewRouter(
		healthcare.NewPatientService(),
	)
	httpServer := httptest.NewServer(router)
	t.Cleanup(httpServer.Close)

	session := connectTestClient(
		t,
		httpServer.URL+"/mcp",
		"tenant-demo",
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	tools, err := session.ListTools(
		ctx,
		&mcp.ListToolsParams{},
	)
	if err != nil {
		t.Fatalf("list MCP tools: %v", err)
	}

	// Check the advertised tool without depending on the
	// SDK's internal representation of its tool schema.
	toolJSON, err := json.Marshal(tools)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(toolJSON), "search_patients") {
		t.Fatalf("search_patients not advertised: %s", toolJSON)
	}

	result, err := session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name: "search_patients",
			Arguments: map[string]any{
				"query": "Alex",
				"limit": 10,
			},
		},
	)
	if err != nil {
		t.Fatalf("call search_patients: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned an MCP error: %+v", result)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"pat-001",
		"Alex Morgan",
		"active",
	} {
		if !strings.Contains(string(resultJSON), expected) {
			t.Errorf(
				"result missing %q: %s",
				expected, resultJSON,
			)
		}
	}
}

func TestSearchPatientsCannotCrossTenants(t *testing.T) {
	router := httpserver.NewRouter(
		healthcare.NewPatientService(),
	)
	httpServer := httptest.NewServer(router)
	t.Cleanup(httpServer.Close)

	session := connectTestClient(
		t,
		httpServer.URL+"/mcp",
		"tenant-demo",
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	result, err := session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name: "search_patients",
			Arguments: map[string]any{
				"query": "Taylor",
				"limit": 10,
			},
		},
	)
	if err != nil {
		t.Fatalf("call search_patients: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned an MCP error: %+v", result)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(resultJSON), "pat-003") {
		t.Fatalf(
			"cross-tenant patient leaked in result: %s",
			resultJSON,
		)
	}
}
