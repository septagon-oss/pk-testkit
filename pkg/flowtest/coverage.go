// Package flowtest validates requirement-to-flow coverage.
package flowtest

// Implements: REQ-015.
// Per: ADR-0021.
// Discipline: C-14.
// coverage.go owns adapter-neutral flow coverage validation over pk-shared
// flow definitions.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (file purpose declaration).

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/septagon-oss/pk-shared/pkg/flowdef"
)

// Requirement is the smallest requirement vocabulary needed by the OSS
// conformance layer.
type Requirement struct {
	ID       string
	Title    string
	Critical bool
}

// Diagnostic describes a coverage or flow definition issue.
type Diagnostic struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Report is the deterministic coverage report.
type Report struct {
	Requirements int          `json:"requirements"`
	Flows        int          `json:"flows"`
	Missing      []string     `json:"missing,omitempty"`
	Diagnostics  []Diagnostic `json:"diagnostics,omitempty"`
}

// OK reports whether all requirements are covered and all flows are valid.
func (r Report) OK() bool {
	return len(r.Missing) == 0 && len(r.Diagnostics) == 0
}

// ValidateCoverage verifies that every requirement is fulfilled by at least
// one valid flow definition.
func ValidateCoverage(requirements []Requirement, flows []flowdef.Definition) Report {
	report := Report{Requirements: len(requirements), Flows: len(flows)}

	required := map[string]Requirement{}
	for _, requirement := range requirements {
		id := strings.TrimSpace(requirement.ID)
		if id == "" {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{ID: "requirement", Message: "requirement ID is required"})
			continue
		}
		if _, exists := required[id]; exists {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{ID: id, Message: "duplicate requirement ID"})
			continue
		}
		required[id] = Requirement{ID: id, Title: strings.TrimSpace(requirement.Title), Critical: requirement.Critical}
	}

	covered := map[string]struct{}{}
	seenFlows := map[string]struct{}{}
	for _, flow := range flows {
		id := strings.TrimSpace(flow.ID)
		if id == "" {
			id = "flow"
		}
		if _, exists := seenFlows[id]; exists {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{ID: id, Message: "duplicate flow ID"})
			continue
		}
		seenFlows[id] = struct{}{}
		if err := flow.Validate(); err != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{ID: id, Message: err.Error()})
			continue
		}
		for _, reqID := range flow.Fulfills {
			reqID = strings.TrimSpace(reqID)
			if reqID == "" {
				continue
			}
			if _, ok := required[reqID]; !ok {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{ID: id, Message: fmt.Sprintf("flow fulfills unknown requirement %q", reqID)})
				continue
			}
			covered[reqID] = struct{}{}
		}
	}

	for id := range required {
		if _, ok := covered[id]; !ok {
			report.Missing = append(report.Missing, id)
		}
	}
	slices.Sort(report.Missing)
	slices.SortStableFunc(report.Diagnostics, func(a, b Diagnostic) int {
		if a.ID == b.ID {
			return cmp.Compare(a.Message, b.Message)
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return report
}
