# healthcare-mcp-gateway

# Healthcare MCP Gateway

A secure, multi-tenant Model Context Protocol (MCP) gateway built with Go, designed to make healthcare API capabilities accessible to AI assistants through a small, well-defined set of tools.

**Project status:** In active development  
**Language:** Go  
**MCP transport:** Streamable HTTP  
**Current implementation:** Patient search, bearer-token authentication, tenant-scoped data access, and end-to-end tests

---

## Table of Contents

- [Overview](#overview)
- [Problem Statement](#problem-statement)
- [Project Goals](#project-goals)
- [Current Implementation Status](#current-implementation-status)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [Authentication and Authorization](#authentication-and-authorization)
- [MCP Tools](#mcp-tools)
- [HTTP Endpoints](#http-endpoints)
- [Testing](#testing)
- [Security Considerations](#security-considerations)
- [Production Roadmap](#production-roadmap)
- [Deployment Strategy](#deployment-strategy)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

The Healthcare MCP Gateway is a Go service that exposes healthcare-related functionality through the Model Context Protocol.

MCP allows AI applications and agents to discover and invoke tools through a standardized interface. Instead of exposing every underlying REST API endpoint directly to an AI agent, this project aims to provide a smaller set of task-oriented tools that combine related operations behind a consistent interface.

The gateway is being developed with several important engineering requirements:

- **Small tool surface:** expose a limited number of meaningful MCP tools rather than mirroring every upstream API endpoint.
- **Authentication:** require a valid bearer token for access to the MCP endpoint.
- **Tenant isolation:** restrict patient searches to the tenant identified by the verified token.
- **Testability:** support unit, integration, and end-to-end testing.
- **Maintainability:** separate HTTP handling, authentication, MCP registration, and healthcare business logic.
- **Deployment readiness:** prepare the service for containerization and eventual deployment to AWS.

The current version uses an in-memory dataset containing synthetic patient records. Integration with a real healthcare API, additional MCP tools, production identity-provider integration, and AWS deployment remain future milestones.

## Problem Statement

Healthcare API providers may expose many REST endpoints for patient records, appointments, medication information, and clinical observations.

Giving an AI agent direct access to every endpoint can create an unnecessarily large tool surface and complicate validation, permissions, and maintenance.

This project explores a different approach:

1. Authenticate the incoming request.
2. Establish the identity and tenant associated with that request.
3. Expose task-oriented MCP tools.
4. Validate tool inputs.
5. Execute the relevant healthcare business logic or upstream API operations.
6. Return only the data the tool is intended to expose.

The goal is to create a practical gateway between AI agents and healthcare systems while keeping security boundaries and API responsibilities explicit.

## Project Goals

### Primary goals

- Build an MCP server in Go using the official Go MCP SDK.
- Expose MCP tools through Streamable HTTP.
- Authenticate requests using signed bearer tokens.
- Enforce tenant-scoped data access.
- Keep MCP handlers separate from healthcare business logic.
- Write automated tests for authentication, data isolation, and tool execution.
- Design a smaller MCP interface over a larger underlying healthcare API.

### Longer-term goals

- Integrate with an external healthcare API.
- Add appointment, medication, and observation capabilities.
- Support a configurable OIDC/JWKS-based authentication provider.
- Implement role- and scope-based authorization.
- Add structured audit logging and operational monitoring.
- Package the service with Docker and deploy it to AWS.
- Establish a repeatable CI/CD pipeline.

---

## Current Implementation Status

The following table distinguishes existing functionality from planned work.

| Component | Status | Details |
|---|---|---|
| Go application entry point | Implemented | Loads configuration and starts the HTTP server |
| MCP server | Implemented | Registers `search_patients` |
| Streamable HTTP transport | Implemented | MCP endpoint at `/mcp` |
| Health endpoint | Implemented | Public `/health` endpoint |
| JWT bearer authentication | Implemented | Verifies HS256 tokens, issuer, audience, expiry, subject, and tenant claim |
| Tenant-scoped patient search | Implemented | Searches only records belonging to the authenticated tenant |
| MCP end-to-end tests | Implemented | Connects an MCP client and invokes the patient search tool |
| Authentication tests | Implemented | Checks missing and invalid bearer tokens |
| Race detector and static checks | Passed at last recorded run | `go test -race ./...`, `go vet ./...`, and `go build ./cmd/server` |
| OIDC/JWKS authentication | In progress | Provider-independent asymmetric-token verification is being explored; not yet considered production-verified |
| Additional MCP tools | Planned | Appointments, medication summaries, and patient observations |
| Real healthcare API integration | Planned | Requires provider API documentation and credentials |
| Role/scope authorization | Planned | More granular permissions beyond tenant claims |
| Audit logging and observability | Planned | Structured events, metrics, and operational alerts |
| Docker and AWS deployment | Planned | Production infrastructure and deployment pipeline |

> **Important:** A passing test suite demonstrates the tested behavior, not full production readiness, regulatory compliance, or security certification.

---

## Architecture

The application separates transport, authentication, MCP tool registration, and healthcare business logic.

```text
                 MCP Client / AI Agent
                          |
                          | HTTPS
                          v
              +-----------------------+
              | Healthcare MCP Gateway|
              |                       |
              |  HTTP Router          |
              |  Bearer Authentication|
              +-----------+-----------+
                          |
                          v
              +-----------------------+
              | MCP Streamable HTTP   |
              | Endpoint: /mcp        |
              +-----------+-----------+
                          |
                          v
              +-----------------------+
              | MCP Tool Registry     |
              |                       |
              | search_patients       |
              +-----------+-----------+
                          |
                          v
              +-----------------------+
              | Healthcare Service    |
              |                       |
              | Input validation      |
              | Tenant-scoped search  |
              +-----------+-----------+
                          |
                          v
              +-----------------------+
              | Synthetic Patient Data|
              | In-memory repository  |
              +-----------------------+
```

The current implementation uses synthetic in-memory records. In the intended architecture, the healthcare service layer will call an external healthcare API through a dedicated client.

### Request lifecycle

1. An MCP client sends a request to `/mcp`.
2. The authentication middleware extracts the bearer token.
3. The JWT verifier validates the token.
4. The verified identity and tenant ID are attached to the request context.
5. The MCP handler registers the tool interface for that tenant.
6. The MCP tool handler calls the healthcare service.
7. The service validates its inputs and limits its search to the authenticated tenant.
8. The tool returns the approved patient fields to the MCP client.

The tenant is derived from a verified token, not from a caller-supplied `X-Tenant-ID` header.

---

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Application language |
| Gin | HTTP routing and middleware |
| Official Go MCP SDK | MCP server, tool registration, and transport |
| `github.com/golang-jwt/jwt/v5` | Current HS256 JWT verification |
| `net/http` | HTTP server and request handling |
| Go testing package | Unit and integration testing |
| `httptest` | HTTP-level tests |
| Git and GitHub | Version control and portfolio hosting |
| AWS | Planned hosting environment |

The `lestrrat-go/jwx/v2` library has been explored for OIDC/JWKS-based verification. Its integration should be treated as in progress until it is wired into configuration and the application entry point, compiled, and covered by security-focused tests.

---

## Project Structure

The core files currently implemented are organized as follows:

```text
healthcare-mcp-gateway/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   └── middleware.go
│   │
│   ├── healthcare/
│   │   ├── patients.go
│   │   └── patients_test.go
│   │
│   ├── httpserver/
│   │   ├── router.go
│   │   └── router_test.go
│   │
│   └── mcpserver/
│       └── server.go
│
├── tests/
│   └── e2e/
│       └── patients_test.go
│
├── go.mod
├── go.sum
└── README.md
```

Additional directories, such as `internal/audit/`, `internal/config/`, and `deploy/aws/`, are planned architectural additions. They should be created when their implementations are introduced rather than assumed to exist already.

### Responsibilities

**`cmd/server/main.go`**

Loads environment configuration, validates the authentication configuration, constructs the services and router, and starts the HTTP server.

**`internal/auth/`**

Validates bearer tokens and makes the authenticated identity available to downstream handlers.

**`internal/healthcare/`**

Contains healthcare domain types, input validation, and tenant-scoped business logic. The current patient service uses synthetic in-memory data.

**`internal/httpserver/`**

Configures HTTP routes, middleware, the health endpoint, and the MCP Streamable HTTP handler.

**`internal/mcpserver/`**

Defines MCP tool inputs and outputs and registers tools with the official Go MCP SDK.

**`tests/e2e/`**

Exercises the MCP interface through a client connection, including authentication and tenant-isolation scenarios.

---

## Getting Started

### Prerequisites

Install the following tools:

- Go 1.24 or a compatible newer version.
- Git.
- `openssl` for generating a local development secret.
- A terminal or editor such as VS Code.

Check your Go installation:

```bash
go version
```

The project has been tested with Go `1.24.4 linux/amd64`.

### Clone the repository

```bash
git clone https://github.com/Zolet-hash/healthcare-mcp-gateway.git

cd healthcare-mcp-gateway
```

### Download dependencies

```bash
go mod download
```

Verify dependency metadata:

```bash
go mod tidy
git status
```

If `go mod tidy` changes `go.mod` or `go.sum`, review and commit those changes only when they are intentional.

---

## Configuration

The current application expects these environment variables:

| Variable | Required | Description |
|---|---|---|
| `JWT_SECRET` | Yes | Shared secret used to verify HS256 tokens |
| `JWT_ISSUER` | Yes | Expected token issuer |
| `JWT_AUDIENCE` | Yes | Expected token audience |
| `PORT` | No | HTTP listening port; defaults to `8080` |

### Configure local development

Generate a random secret and export the configuration:

```bash
export JWT_SECRET="$(openssl rand -base64 32)"
export JWT_ISSUER="healthcare-local-dev"
export JWT_AUDIENCE="healthcare-mcp-gateway"
export PORT=8080
```

The secret should remain in your local environment. Do not commit secrets to GitHub or put production credentials in the README.

The local secret must be available to the component that signs test or development tokens as well as to the server that verifies them.

**Note:** The application does not currently expose a login endpoint or issue tokens itself. A valid token must be issued by a compatible trusted token issuer or generated by a controlled development/test utility.

The issuer, audience, signing algorithm, expiry, subject, and `tenant_id` claim must satisfy the verifier's requirements.

---

## Running the Server

With the environment variables configured, start the application:

```bash
go run ./cmd/server
```

The server listens on the configured port, defaulting to `8080`.

### Verify the health endpoint

In another terminal:

```bash
curl -i http://localhost:8080/health
```

Expected response:

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
```

Response body:

```json
{
  "status": "ok"
}
```

The health endpoint is intentionally public in the current implementation. It indicates that the HTTP process is responding; it does not establish that a future upstream healthcare API or database is healthy.

### Check the MCP endpoint

The MCP endpoint is:

```text
http://localhost:8080/mcp
```

It requires a valid bearer token.

A request without credentials should receive an HTTP `401 Unauthorized` response:

```bash
curl -i http://localhost:8080/mcp
```

The MCP endpoint is intended for MCP-compatible clients. A simple `curl` request does not replace a complete MCP initialization and tool-call exchange.

---

## Authentication and Authorization

### Current authentication model

The current implementation uses signed JWT bearer tokens.

A verified token must satisfy the configured requirements, including:

- An accepted signing algorithm: HS256 in the current verifier.
- A valid signature.
- The expected issuer.
- The expected audience.
- A valid expiration time, with the configured clock-skew allowance.
- A non-empty subject identifying the user.
- A non-empty `tenant_id` claim.

The application converts the validated claims into an internal identity:

```go
type Identity struct {
    UserID   string
    TenantID string
}
```

The identity is then made available to downstream HTTP and MCP request handling.

### Tenant isolation

Patient records are associated with tenant IDs.

The current synthetic dataset contains records for two tenants:

| Tenant | Patient ID | Display name | Status |
|---|---|---|---|
| `tenant-demo` | `pat-001` | Alex Morgan | Active |
| `tenant-demo` | `pat-002` | Jordan Lee | Inactive |
| `tenant-other` | `pat-003` | Taylor Reed | Active |

These names and records are synthetic and intended only for development and testing.

When a request is authenticated as `tenant-demo`, patient search must not return `pat-003`, even if the caller sends:

```http
X-Tenant-ID: tenant-other
```

The server uses the verified identity rather than trusting that header.

### Current limitations

A valid signed tenant claim establishes which tenant the token asserts. It does not, by itself, prove that the user should be authorized to access that tenant.

Before production, the authorization design should address:

- User membership in a tenant.
- Roles and scopes.
- Permissions for individual tools and data categories.
- Token revocation and identity-provider policies.
- Signing-key rotation.
- The selected healthcare provider's identity and access model.

### OIDC and JWKS support

The next authentication milestone is a configurable asymmetric-signature verifier using a trusted provider's JWKS endpoint.

This work should include issuer and audience validation, algorithm restrictions, key rotation, rejection of unknown keys, expiry validation, and automated tests. Provider-specific identity configuration must be supplied before this integration can be considered complete.

The existing HS256 implementation is suitable for the current controlled development and test setup; it should not automatically be assumed to be the correct production identity model.

---

## MCP Tools

### Implemented tool

#### `search_patients`

Searches for synthetic patient records belonging to the authenticated tenant.

**Input:**

```json
{
  "query": "Alex",
  "limit": 10
}
```

The input contains:

| Field | Purpose |
|---|---|
| `query` | Patient-name fragment to search for; must contain at least two non-whitespace characters |
| `limit` | Maximum number of results, from 1 to 20 |

**Output example:**

```json
{
  "patients": [
    {
      "id": "pat-001",
      "display_name": "Alex Morgan",
      "status": "active"
    }
  ],
  "count": 1
}
```

The actual MCP response wraps the tool result according to the MCP SDK's response format.

The current tool:

- Validates search input.
- Enforces a result limit.
- Searches only within the authenticated tenant.
- Returns a limited set of patient fields.
- Does not query a real healthcare API or production database.

### Planned MCP interface

The intended architecture is to expose four high-level tools over approximately twelve underlying healthcare REST API operations. The exact REST endpoint mapping will depend on the provider's actual API contract.

| MCP tool | Intended responsibility | Status |
|---|---|---|
| `search_patients` | Search for patients within the caller's authorized tenant | Implemented against synthetic data |
| `manage_appointments` | Support authorized appointment-related operations | Planned |
| `get_medication_summary` | Retrieve a concise medication summary | Planned |
| `get_patient_observations` | Retrieve relevant clinical observations | Planned |

The four-tool design is intended to reduce the number of agent-facing tools while preserving the underlying API's functionality. It does not mean the three planned tools or all twelve REST operations have already been implemented.

Tool descriptions, schemas, and permissions will need to be designed carefully so that an AI agent cannot use a broad tool to bypass the user's authorization.

---

## HTTP Endpoints

| Method | Path | Authentication | Purpose |
|---|---|---|---|
| `GET` | `/health` | Not required | Basic process health check |
| `GET`, `POST`, and other methods supported by the registered handler | `/mcp` | Bearer token required | MCP Streamable HTTP transport |

The MCP handler is mounted on `/mcp`; the official MCP SDK manages the MCP protocol behavior.

The gateway is not currently a general-purpose REST API for patient CRUD operations. Healthcare functionality is exposed through MCP tools.

---

## Testing

Automated tests are an important part of the project, especially because cross-tenant data exposure would be a serious security issue.

### Run all tests

```bash
go test ./...
```

### Run tests with the race detector

```bash
go test -race ./...
```

### Run static analysis

```bash
go vet ./...
```

### Build the server

```bash
go build ./cmd/server
```

This command builds the server executable in the current directory. If you want to build without leaving an executable in the repository root, you can use:

```bash
go build -o /tmp/healthcare-mcp-gateway ./cmd/server
```

### Run focused tests

Authentication:

```bash
go test ./internal/auth -v
```

Healthcare service:

```bash
go test ./internal/healthcare -v
```

HTTP server:

```bash
go test ./internal/httpserver -v
```

End-to-end MCP tests:

```bash
go test ./tests/e2e -v
```

### Existing test coverage

The current test suite covers:

- Patient search within a tenant.
- Rejection of invalid patient search input.
- Public health endpoint behavior.
- Rejection of requests without bearer credentials.
- Rejection of invalid bearer tokens.
- MCP client connection and patient-search tool invocation.
- Prevention of cross-tenant patient exposure when a caller attempts to spoof `X-Tenant-ID`.

The last recorded validation run completed successfully for:

```bash
gofmt -w internal/auth internal/httpserver cmd/server tests/e2e
go test -race ./...
go vet ./...
go build ./cmd/server
```

Run these commands again after making changes. A previous successful run does not guarantee that subsequent changes still pass.

### Tests to add

Before production deployment, the test suite should also cover:

- Expired tokens and malformed claims.
- Incorrect issuer and audience.
- Token-signature failures.
- JWKS retrieval failures and signing-key rotation.
- Role- and scope-based authorization.
- Every MCP tool's input and output contract.
- Upstream API timeouts, rate limits, and error responses.
- Context cancellation and concurrency.
- Redaction of sensitive data in logs.
- Authorization checks on every patient-specific operation.

---

## Security Considerations

Healthcare systems handle sensitive information, so security must be part of the architecture rather than an afterthought.

### Current safeguards

- Bearer authentication protects the MCP endpoint.
- JWT validation checks the expected issuer, audience, expiry, subject, and tenant claim.
- Tenant context comes from verified token claims rather than an untrusted request header.
- Patient search validates its inputs and limits result counts.
- The current dataset is synthetic.
- Authentication errors avoid exposing detailed token-validation internals.

### Production requirements

Before using real healthcare data, the project needs a broader security review.

**Transport security**

- Serve external traffic over HTTPS.
- Configure TLS termination and secure upstream connections.
- Restrict network access to internal services.

**Identity and authorization**

- Integrate with the selected identity provider.
- Define tenant membership and per-tool permissions.
- Verify signing-key rotation and revocation requirements.
- Fail closed if authentication or authorization cannot be established.

**Data protection**

- Use synthetic data for public demonstrations.
- Avoid returning unnecessary patient attributes.
- Redact bearer tokens and sensitive patient information from logs.
- Define retention and deletion policies.
- Protect credentials and encryption keys.

**Reliability**

- Configure upstream timeouts.
- Use safe retry policies, accounting for whether operations are idempotent.
- Handle pagination and rate limits.
- Validate external API responses.
- Avoid silently converting upstream failures into successful empty results.

**Auditability**

- Record appropriate security and tool-execution events.
- Include request correlation IDs, actor identity, tenant, tool name, and outcome where appropriate.
- Avoid logging complete patient records or credentials.
- Define access controls and retention for audit logs.

This repository should not be described as HIPAA-compliant or compliant with any particular data-protection framework without an appropriate legal, technical, organizational, and operational assessment.

---

## Production Roadmap

The work is divided into milestones so that each layer can be developed and validated independently.

### Milestone 1: Authentication hardening

- Complete configurable OIDC/JWKS verification.
- Add asymmetric-signature and key-rotation tests.
- Integrate the selected provider configuration into application startup.
- Ensure invalid authentication fails closed.

### Milestone 2: Authorization and tenant isolation

- Define tenant membership, roles, and scopes.
- Enforce permissions for each MCP tool.
- Test unauthorized cross-tenant and cross-resource access.
- Document the authorization model.

### Milestone 3: Healthcare API integration

- Obtain the upstream API documentation.
- Implement an HTTP client with explicit timeouts.
- Handle upstream authentication and credential rotation.
- Map approximately twelve REST operations into four task-oriented MCP tools.
- Validate request and response schemas.
- Add mocked and integration tests.

### Milestone 4: Reliability and API contracts

- Standardize errors.
- Implement safe retry and idempotency behavior where appropriate.
- Add pagination and rate-limit handling.
- Support context cancellation.
- Test upstream service failures.

### Milestone 5: Audit logging and data protection

- Introduce structured audit events.
- Redact sensitive information.
- Define retention and access policies.
- Secure secrets and encryption keys.
- Document data flows and privacy controls.

### Milestone 6: Security and automated testing

- Expand unit and end-to-end test coverage.
- Test authorization boundaries.
- Add dependency and vulnerability scanning.
- Run security-focused reviews.
- Document known limitations and threat assumptions.

### Milestone 7: Containerization and AWS deployment

- Add a production Dockerfile.
- Configure environment-specific settings.
- Use an AWS secrets-management solution.
- Apply least-privilege IAM policies.
- Configure HTTPS, networking, health checks, and resource limits.
- Establish deployment and rollback procedures.

### Milestone 8: Observability and operational readiness

- Add structured application logs and useful metrics.
- Configure alerts for failures and latency.
- Monitor upstream API availability and rate limits.
- Define incident-response procedures.
- Document backups and recovery where applicable.
- Test deployment and rollback workflows.

The project should be considered production-ready only after these requirements have been assessed against the actual provider, deployment environment, data sensitivity, and intended use.

---

## Deployment Strategy

AWS is the intended deployment target, but deployment infrastructure is not yet complete.

A possible future deployment architecture is:

```text
MCP Client
    |
    | HTTPS
    v
AWS Load Balancer / TLS
    |
    v
Go MCP Gateway
    |
    +----> Identity Provider / JWKS
    |
    +----> Healthcare API Provider
    |
    +----> Audit Logging and Monitoring
```

The final architecture will depend on the provider's networking requirements, identity model, availability targets, and data-handling constraints.

Production configuration should use managed secrets, least-privilege IAM, restricted network access, health checks, monitoring, and an explicit deployment strategy.

Do not deploy real patient data or production credentials into the public demo environment.

---

## Contributing

Contributions and technical feedback are welcome.

For changes to this repository:

1. Create a focused branch.
2. Make the smallest change that implements the intended behavior.
3. Add or update tests.
4. Format the code.
5. Run the complete validation suite.
6. Review the diff for secrets and sensitive data.
7. Open a pull request describing the change and its tests.

Before submitting changes, run:

```bash
gofmt -w ./cmd ./internal ./tests
go test -race ./...
go vet ./...
go build ./cmd/server
```

If new packages or tools are added, update this README to reflect their actual implementation status.

---

## License

No license has been specified yet. Add a `LICENSE` file before presenting this project as open source or granting others permission to reuse its code.

---

## Disclaimer

This is a software engineering project intended to demonstrate Go development, MCP integration, API gateway design, authentication, tenant isolation, and automated testing.

The current implementation uses synthetic patient data. It is not a substitute for a clinically validated healthcare system, a production security assessment, or a regulatory compliance review.
