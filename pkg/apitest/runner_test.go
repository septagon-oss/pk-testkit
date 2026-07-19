package apitest

// runner_test.go validates API flow execution against in-process HTTP handlers.
//
// Validates: REQ-015.
// Per: ADR-0029 (file purpose declaration).
// Discipline: C-14.

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

func TestRunnerStopsWhenRequestBuilderCancelsContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	handlerCalled := false
	runner, err := NewRunner(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusNoContent)
		}),
		WithRequestBuilder(func(ctx context.Context, definition flowdef.Definition, step flowdef.APIStep) (*http.Request, error) {
			cancel()
			return DefaultRequestBuilder(ctx, definition, step)
		}),
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	result := runner.Run(ctx, flowdef.Definition{
		ID:   "cancel-during-build",
		Name: "Cancel During Build",
		Channels: flowdef.Channels{API: &flowdef.APIChannel{Steps: []flowdef.APIStep{{
			OperationID:     "cancel",
			Method:          "GET",
			Path:            "/cancel",
			SuccessStatuses: []int{http.StatusNoContent},
		}}}},
	})
	if result.Passed {
		t.Fatalf("flow should fail after cancellation: %#v", result)
	}
	if handlerCalled {
		t.Fatal("handler ran after request builder canceled the context")
	}
	if len(result.Steps) != 1 || result.Steps[0].Message != context.Canceled.Error() {
		t.Fatalf("canceled step = %#v", result.Steps)
	}
}
