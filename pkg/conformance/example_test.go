package conformance_test

// example_test.go provides runnable godoc examples for the conformance package.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"context"
	"fmt"

	"github.com/septagon-oss/pk-testkit/pkg/conformance"
)

// ExampleSuite_Run builds a one-check suite and prints the aggregated status.
func ExampleSuite_Run() {
	suite, err := conformance.NewSuite(conformance.Check{
		ID:            "runtime.ready",
		RequirementID: "REQ-READY",
		Description:   "runtime reports ready",
		Run: func(context.Context) (conformance.Result, error) {
			return conformance.Pass("runtime is ready"), nil
		},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	report := suite.Run(context.Background())
	fmt.Println(report.Status)
	fmt.Println(report.OK())
	// Output:
	// pass
	// true
}
