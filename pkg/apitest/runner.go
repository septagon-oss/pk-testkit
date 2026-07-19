// Package apitest executes API flow definitions against an http.Handler.
package apitest

// runner.go owns a standard-library API flow runner. It deliberately avoids
// network clients and environment orchestration so the same flow definitions can
// run in unit tests, local apps, or downstream E2E harnesses.
//
// Implements: REQ-015.
// Per: ADR-0029 (file purpose declaration).
// Discipline: C-14.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"time"

	"github.com/septagon-oss/pk-shared/pkg/flowdef"
)

// RequestBuilder builds one request for a flow step.
type RequestBuilder func(context.Context, flowdef.Definition, flowdef.APIStep) (*http.Request, error)

// Runner executes API flows over an HTTP handler.
type Runner struct {
	handler http.Handler
	builder RequestBuilder
}

// Option configures a Runner.
type Option func(*Runner)

// WithRequestBuilder overrides default request construction.
func WithRequestBuilder(builder RequestBuilder) Option {
	return func(r *Runner) {
		if builder != nil {
			r.builder = builder
		}
	}
}

// NewRunner returns an API flow runner.
func NewRunner(handler http.Handler, opts ...Option) (*Runner, error) {
	if handler == nil {
		return nil, fmt.Errorf("apitest runner handler is required")
	}
	runner := &Runner{handler: handler, builder: DefaultRequestBuilder}
	for _, opt := range opts {
		if opt != nil {
			opt(runner)
		}
	}
	return runner, nil
}

// FlowResult is the result of one flow execution.
type FlowResult struct {
	ID       string        `json:"id"`
	Passed   bool          `json:"passed"`
	Message  string        `json:"message,omitempty"`
	Duration time.Duration `json:"duration"`
	Steps    []StepResult  `json:"steps"`
}

// StepResult is the result of one API step.
type StepResult struct {
	OperationID string        `json:"operation_id"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	StatusCode  int           `json:"status_code"`
	Passed      bool          `json:"passed"`
	Message     string        `json:"message,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// Run executes the API channel for one flow definition.
func (r *Runner) Run(ctx context.Context, definition flowdef.Definition) FlowResult {
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()
	result := FlowResult{ID: strings.TrimSpace(definition.ID), Passed: true}
	if r == nil || r.handler == nil {
		result.Passed = false
		result.Message = "apitest runner handler is required"
		result.Duration = time.Since(start)
		return result
	}
	if err := definition.Validate(); err != nil {
		result.Passed = false
		result.Message = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	if definition.Channels.API == nil {
		result.Passed = false
		result.Message = "flow has no api channel"
		result.Duration = time.Since(start)
		return result
	}

	for _, step := range definition.Channels.API.NormalizedSteps() {
		stepResult := r.runStep(ctx, definition, step)
		if !stepResult.Passed {
			result.Passed = false
		}
		result.Steps = append(result.Steps, stepResult)
	}
	result.Duration = time.Since(start)
	return result
}

func (r *Runner) runStep(ctx context.Context, definition flowdef.Definition, step flowdef.APIStep) StepResult {
	start := time.Now()
	result := StepResult{
		OperationID: strings.TrimSpace(step.OperationID),
		Method:      strings.ToUpper(strings.TrimSpace(step.Method)),
		Path:        strings.TrimSpace(step.Path),
		Passed:      false,
	}
	if err := ctx.Err(); err != nil {
		result.Message = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	req, err := r.builder(ctx, definition, step)
	if err != nil {
		result.Message = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	if err := ctx.Err(); err != nil {
		result.Message = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	rec := httptest.NewRecorder()
	r.handler.ServeHTTP(rec, req)
	result.StatusCode = rec.Code
	result.Passed = successStatus(step.SuccessStatuses, rec.Code)
	if !result.Passed {
		result.Message = fmt.Sprintf("status %d did not match expected statuses %v", rec.Code, expectedStatuses(step.SuccessStatuses))
	}
	result.Duration = time.Since(start)
	return result
}

// DefaultRequestBuilder constructs a no-body HTTP request for the step path and
// copies declared static headers.
func DefaultRequestBuilder(ctx context.Context, _ flowdef.Definition, step flowdef.APIStep) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := step.Validate(); err != nil {
		return nil, err
	}
	req := httptest.NewRequest(strings.ToUpper(strings.TrimSpace(step.Method)), strings.TrimSpace(step.Path), nil)
	req = req.WithContext(ctx)
	for key, value := range step.Headers {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		req.Header.Set(key, strings.TrimSpace(value))
	}
	return req, nil
}

func successStatus(expected []int, got int) bool {
	if len(expected) == 0 {
		return got >= 200 && got <= 299
	}
	return slices.Contains(expected, got)
}

func expectedStatuses(statuses []int) []int {
	if len(statuses) == 0 {
		return []int{200, 201, 202, 203, 204, 205, 206}
	}
	return slices.Clone(statuses)
}
