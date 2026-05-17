package apitest

// runner_test.go validates API flow execution against in-process HTTP handlers.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"context"
	"net/http"
	"testing"

	"github.com/septagon-oss/pk-shared/pkg/flowdef"
)

func TestRunnerExecutesAPIFlow(t *testing.T) {
	t.Parallel()

	runner, err := NewRunner(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ready" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	result := runner.Run(context.Background(), flowdef.Definition{
		ID:   "ready",
		Name: "Ready",
		Channels: flowdef.Channels{API: &flowdef.APIChannel{Steps: []flowdef.APIStep{{
			OperationID:     "ready",
			Method:          "GET",
			Path:            "/ready",
			SuccessStatuses: []int{204},
		}}}},
	})
	if !result.Passed {
		t.Fatalf("flow should pass: %#v", result)
	}
}

func TestRunnerFailsUnexpectedStatus(t *testing.T) {
	t.Parallel()

	runner, err := NewRunner(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	result := runner.Run(context.Background(), flowdef.Definition{
		ID:   "failing",
		Name: "Failing",
		Channels: flowdef.Channels{API: &flowdef.APIChannel{Steps: []flowdef.APIStep{{
			OperationID: "failing",
			Method:      "GET",
			Path:        "/failing",
		}}}},
	})
	if result.Passed {
		t.Fatal("expected flow to fail")
	}
}

func TestRunnerAcceptsNilContext(t *testing.T) {
	t.Parallel()

	runner, err := NewRunner(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	var ctx context.Context
	result := runner.Run(ctx, flowdef.Definition{
		ID:   "nil-context",
		Name: "Nil Context",
		Channels: flowdef.Channels{API: &flowdef.APIChannel{Steps: []flowdef.APIStep{{
			OperationID: "ok",
			Method:      "GET",
			Path:        "/ok",
		}}}},
	})
	if !result.Passed {
		t.Fatalf("flow should pass: %#v", result)
	}
}
