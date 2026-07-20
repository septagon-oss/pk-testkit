package flowtest

// Validates: REQ-015.
// Per: ADR-0021.
// Discipline: C-14.
// coverage_test.go validates requirement-to-flow coverage diagnostics.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"testing"

	"github.com/septagon-oss/pk-shared/pkg/flowdef"
)

func TestValidateCoveragePassesCoveredRequirements(t *testing.T) {
	t.Parallel()

	report := ValidateCoverage(
		[]Requirement{{ID: "REQ-1", Critical: true}},
		[]flowdef.Definition{{
			ID:       "flow-1",
			Name:     "Flow",
			Fulfills: []string{"REQ-1"},
			Channels: flowdef.Channels{API: &flowdef.APIChannel{Steps: []flowdef.APIStep{{
				OperationID:     "getThing",
				Method:          "GET",
				Path:            "/thing",
				SuccessStatuses: []int{200},
			}}}},
		}},
	)
	if !report.OK() {
		t.Fatalf("report should pass: %#v", report)
	}
}

func TestValidateCoverageFailsMissingRequirement(t *testing.T) {
	t.Parallel()

	report := ValidateCoverage([]Requirement{{ID: "REQ-1"}}, nil)
	if report.OK() {
		t.Fatal("expected missing requirement")
	}
	if len(report.Missing) != 1 || report.Missing[0] != "REQ-1" {
		t.Fatalf("missing = %#v", report.Missing)
	}
}
