package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/auth"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/httpserver"
	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func testAuthConfig() auth.Config {
	return auth.Config{
		Secret:   []byte("e2e-test-secret-at-least-32-bytes-long"),
		Issuer:   "healthcare-e2e-test",
		Audience: "healthcare-mcp-gateway",
	}
}

func createTestToken(t *testing.T, cfg auth.Config, tenantID string) string {
	t.Helper()

	now := time.Now()

	claims := auth.Claims{
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "e2e-user",
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(cfg.Secret)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}

	return signed
}

type authTransport struct {
	base        http.RoundTripper
	token       string
	spoofTenant string
}

func (t authTransport) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()

	if t.token != "" {
		cloned.Header.Set("Authorization", "Bearer "+t.token)
	}

	// This header must never override the signed JWT tenant.
	if t.spoofTenant != "" {
		cloned.Header.Set("X-Tenant-ID", t.spoofTenant)
	}

	return t.base.RoundTrip(cloned)
}

func connectTestClient(
	t *testing.T,
	endpoint string,
	cfg auth.Config,
	tenantID string,
	spoofTenant string,
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
		Transport: authTransport{
			base:        http.DefaultTransport,
			token:       createTestToken(t, cfg, tenantID),
			spoofTenant: spoofTenant,
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

func TestMCPRejectsMissingBearerToken(t *testing.T) {
	cfg := testAuthConfig()
	router := httpserver.NewRouter(
		healthcare.NewPatientService(),
		cfg,
	)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	req, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/mcp",
		strings.NewReader("{}"),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("send unauthenticated request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusUnauthorized,
		)
	}
}

func TestSearchPatientsThroughMCP(t *testing.T) {
	cfg := testAuthConfig()
	router := httpserver.NewRouter(
		healthcare.NewPatientService(),
		cfg,
	)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	session := connectTestClient(
		t,
		server.URL+"/mcp",
		cfg,
		"tenant-demo",
		"",
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list MCP tools: %v", err)
	}

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
			t.Errorf("result missing %q: %s", expected, resultJSON)
		}
	}
}

func TestSearchPatientsCannotCrossTenants(t *testing.T) {
	cfg := testAuthConfig()
	router := httpserver.NewRouter(
		healthcare.NewPatientService(),
		cfg,
	)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	// The signed token belongs to tenant-demo, but the client
	// attempts to spoof tenant-other using X-Tenant-ID.
	session := connectTestClient(
		t,
		server.URL+"/mcp",
		cfg,
		"tenant-demo",
		"tenant-other",
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
		t.Fatalf("cross-tenant patient leaked in result: %s", resultJSON)
	}
}
