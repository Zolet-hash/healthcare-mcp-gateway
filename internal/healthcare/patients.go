
package healthcare

import (
	"context"
	"fmt"
	"strings"
)

type Patient struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}

type PatientService struct {
	patients []struct {
		TenantID string
		Patient  Patient
	}
}

func NewPatientService() *PatientService {
	s := &PatientService{} //pointer

	s.patients = append(s.patients,
		struct {
			TenantID string
			Patient  Patient
		}{
			TenantID: "tenant-demo",
			Patient: Patient{
				ID: "pat-001", DisplayName: "Alex Morgan",
				Status: "active",
			},
		},
		struct {
			TenantID string
			Patient  Patient
		}{
			TenantID: "tenant-demo",
			Patient: Patient{
				ID: "pat-002", DisplayName: "Jordan Lee",
				Status: "inactive",
			},
		},
		struct {
			TenantID string
			Patient  Patient
		}{
			TenantID: "tenant-other",
			Patient: Patient{
				ID: "pat-003", DisplayName: "Kipchoge Keino",
				Status: "active",
			},
		},
	)

	return s
}

func (s *PatientService) Search(
	ctx context.Context,
	tenantID string,
	query string,
	limit int,
) ([]Patient, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant identity is required")
	}
	if len(strings.TrimSpace(query)) < 2 {
		return nil, fmt.Errorf("query must contain at least 2 characters")
	}
	if limit < 1 || limit > 20 {
		return nil, fmt.Errorf("limit must be between 1 and 20")
	}

	query = strings.ToLower(strings.TrimSpace(query))
	results := make([]Patient, 0)

	for _, record := range s.patients {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Enforce tenant isolation before returning any patient.
		if record.TenantID != tenantID {
			continue
		}

		name := strings.ToLower(record.Patient.DisplayName)
		if strings.Contains(name, query) {
			results = append(results, record.Patient)
			if len(results) == limit {
				break
			}
		}
	}

	return results, nil
}
