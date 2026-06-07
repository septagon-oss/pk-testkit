// Package conformance runs deterministic platform conformance checks.
package conformance

// conformance.go owns the adapter-neutral conformance check contract for OSS
// and downstream PlatformKit distributions.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (file purpose declaration).

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Status is the normalized outcome for one check.
type Status string

const (
	// StatusPass marks a check that satisfied its requirement.
	StatusPass Status = "pass"
	// StatusFail marks a check that did not satisfy its requirement.
	StatusFail Status = "fail"
	// StatusSkip marks a check that was intentionally not evaluated.
	StatusSkip Status = "skip"
)

// Result is one check result.
type Result struct {
	Status  Status            `json:"status"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// Pass returns a passing result.
func Pass(message string) Result {
	return Result{Status: StatusPass, Message: strings.TrimSpace(message)}
}

// Fail returns a failing result.
func Fail(message string) Result {
	return Result{Status: StatusFail, Message: strings.TrimSpace(message)}
}

// Skip returns a skipped result.
func Skip(message string) Result {
	return Result{Status: StatusSkip, Message: strings.TrimSpace(message)}
}

// Check is one deterministic conformance check.
type Check struct {
	ID            string
	RequirementID string
	Description   string
	Run           func(context.Context) (Result, error)
}

// Normalize validates and returns a copy of the check.
func (c Check) Normalize() (Check, error) {
	c.ID = strings.TrimSpace(c.ID)
	c.RequirementID = strings.TrimSpace(c.RequirementID)
	c.Description = strings.TrimSpace(c.Description)
	if c.ID == "" {
		return Check{}, fmt.Errorf("conformance check ID is required")
	}
	if strings.ContainsAny(c.ID, " \t\n\r") {
		return Check{}, fmt.Errorf("conformance check ID %q must not contain whitespace", c.ID)
	}
	if c.RequirementID == "" {
		return Check{}, fmt.Errorf("conformance check %q requirement ID is required", c.ID)
	}
	if c.Run == nil {
		return Check{}, fmt.Errorf("conformance check %q run function is required", c.ID)
	}
	return c, nil
}

// Suite is an immutable set of conformance checks.
type Suite struct {
	checks []Check
}

// NewSuite validates checks and stores them in deterministic order.
func NewSuite(checks ...Check) (*Suite, error) {
	normalized := make([]Check, 0, len(checks))
	seen := map[string]struct{}{}
	for _, check := range checks {
		c, err := check.Normalize()
		if err != nil {
			return nil, err
		}
		if _, exists := seen[c.ID]; exists {
			return nil, fmt.Errorf("duplicate conformance check ID %q", c.ID)
		}
		seen[c.ID] = struct{}{}
		normalized = append(normalized, c)
	}
	slices.SortStableFunc(normalized, func(a, b Check) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return &Suite{checks: normalized}, nil
}

// Run executes all checks in deterministic order.
func (s *Suite) Run(ctx context.Context) Report {
	if ctx == nil {
		ctx = context.Background()
	}
	report := Report{StartedAt: time.Now().UTC(), Status: StatusPass}
	if s == nil || len(s.checks) == 0 {
		report.Status = StatusFail
		report.Results = append(report.Results, CheckResult{
			ID:      "suite.empty",
			Status:  StatusFail,
			Message: "conformance suite has no checks",
		})
		report.CompletedAt = time.Now().UTC()
		return report
	}

	for _, check := range s.checks {
		start := time.Now()
		result := CheckResult{
			ID:            check.ID,
			RequirementID: check.RequirementID,
			Description:   check.Description,
			Status:        StatusFail,
		}
		if err := ctx.Err(); err != nil {
			result.Message = err.Error()
			result.Duration = time.Since(start)
			report.Results = append(report.Results, result)
			continue
		}
		out, err := check.Run(ctx)
		if err != nil {
			result.Message = err.Error()
			result.Duration = time.Since(start)
			report.Results = append(report.Results, result)
			continue
		}
		status, ok := normalizeStatus(out.Status)
		if !ok {
			result.Message = fmt.Sprintf("invalid conformance status %q", out.Status)
		} else {
			result.Status = status
			result.Message = strings.TrimSpace(out.Message)
			result.Details = cloneMap(out.Details)
		}
		result.Duration = time.Since(start)
		report.Results = append(report.Results, result)
	}
	report.CompletedAt = time.Now().UTC()
	report.Status = aggregate(report.Results)
	return report
}

// CheckResult is the executed result of one check.
type CheckResult struct {
	ID            string            `json:"id"`
	RequirementID string            `json:"requirement_id"`
	Description   string            `json:"description,omitempty"`
	Status        Status            `json:"status"`
	Message       string            `json:"message,omitempty"`
	Duration      time.Duration     `json:"duration"`
	Details       map[string]string `json:"details,omitempty"`
}

// Report is the suite execution report.
type Report struct {
	Status      Status        `json:"status"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Results     []CheckResult `json:"results"`
}

// OK reports whether the suite passed.
func (r Report) OK() bool {
	return r.Status == StatusPass
}

func normalizeStatus(status Status) (Status, bool) {
	switch Status(strings.TrimSpace(string(status))) {
	case StatusPass:
		return StatusPass, true
	case StatusFail:
		return StatusFail, true
	case StatusSkip:
		return StatusSkip, true
	default:
		return "", false
	}
}

func aggregate(results []CheckResult) Status {
	if len(results) == 0 {
		return StatusFail
	}
	skipped := 0
	for _, result := range results {
		if result.Status == StatusFail {
			return StatusFail
		}
		if result.Status == StatusSkip {
			skipped++
		}
	}
	if skipped == len(results) {
		return StatusSkip
	}
	return StatusPass
}

func cloneMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	return out
}
