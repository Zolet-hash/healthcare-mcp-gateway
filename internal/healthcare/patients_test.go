package healthcare

import (
	"context"
	"testing"
)

func TestSearchIsTenantScoped(t *testing.T) {
	service := NewPatientService()

	got, err := service.Search(
		context.Background(),
		"tenant-demo",
		"Alex",
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 patient, got %d", len(got))
	}
	if got[0].ID != "pat-001" {
		t.Fatalf("unexpected patient ID: %s", got[0].ID)
	}

	other, err := service.Search(
		context.Background(),
		"tenant-other",
		"Alex",
		10,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Fatalf("expected no cross-tenant results, got %d", len(other))
	}
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	service := NewPatientService()

	tests := []struct {
		name  string
		query string
		limit int
	}{
		{"short query", "A", 10},
		{"zero limit", "Alex", 0},
		{"excessive limit", "Alex", 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Search(
				context.Background(),
				"tenant-demo",
				tt.query,
				tt.limit,
			)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
