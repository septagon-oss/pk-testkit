package conformance

// conformance_test.go validates deterministic conformance check execution and
// validation.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"context"
	"slices"
	"testing"
)

func TestSuiteRunsChecksInDeterministicOrder(t *testing.T) {
	t.Parallel()

	suite, err := NewSuite(
		Check{ID: "z", RequirementID: "REQ-Z", Run: func(context.Context) (Result, error) { return Pass("z"), nil }},
		Check{ID: "a", RequirementID: "REQ-A", Run: func(context.Context) (Result, error) { return Pass("a"), nil }},
	)
	if err != nil {
		t.Fatalf("NewSuite() error = %v", err)
	}

	report := suite.Run(context.Background())
	if !report.OK() {
		t.Fatalf("report status = %q", report.Status)
	}
	got := []string{report.Results[0].ID, report.Results[1].ID}
	if !slices.Equal(got, []string{"a", "z"}) {
		t.Fatalf("check order = %v", got)
	}
}

func TestSuiteFailsWhenEmpty(t *testing.T) {
	t.Parallel()

	report := (*Suite)(nil).Run(context.Background())
	if report.Status != StatusFail {
		t.Fatalf("report.Status = %q, want %q", report.Status, StatusFail)
	}
}

func TestSuiteAcceptsNilContext(t *testing.T) {
	t.Parallel()

	suite, err := NewSuite(Check{ID: "ready", RequirementID: "REQ-READY", Run: func(context.Context) (Result, error) {
		return Pass("ready"), nil
	}})
	if err != nil {
		t.Fatalf("NewSuite() error = %v", err)
	}

	var ctx context.Context
	report := suite.Run(ctx)
	if !report.OK() {
		t.Fatalf("report status = %q", report.Status)
	}
}
