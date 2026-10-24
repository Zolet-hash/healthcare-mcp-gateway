package mcpserver

import (
	"context"

	"github.com/Zolet-hash/healthcare-mcp-gateway/internal/healthcare"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchPatientsInput struct {
	Query string `json:"query" jsonschema:"Name fragment to search for"`
	Limit int    `json:"limit" jsonschema:"Maximum results, from 1 to 20"`
}

type SearchPatientsOutput struct {
	Patients []healthcare.Patient `json:"patients"`
	Count    int                  `json:"count"`
}

func New(
	patients *healthcare.PatientService,
	tenantID string,
) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "healthcare-mcp-gateway",
			Version: "0.1.0",
		},
		nil,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "search_patients",
			Description: "Search synthetic patient records " +
				"within the authenticated tenant. " +
				"Returns only approved patient fields.",
		},
		func(
			ctx context.Context,
			req *mcp.CallToolRequest,
			input SearchPatientsInput,
		) (*mcp.CallToolResult, SearchPatientsOutput, error) {
			found, err := patients.Search(
				ctx, tenantID, input.Query, input.Limit,
			)
			if err != nil {
				return nil, SearchPatientsOutput{}, err
			}

			return nil, SearchPatientsOutput{
				Patients: found,
				Count:    len(found),
			}, nil
		},
	)

	return server
}
