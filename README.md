# pk-testkit

> Part of [PlatformKit](https://github.com/septagon-oss/platformkit) — the open-source Go backend for multi-tenant SaaS.

**Depends on.** `pk-shared` only. Nothing else in PlatformKit.

[![Go Reference](https://pkg.go.dev/badge/github.com/septagon-oss/pk-testkit.svg)](https://pkg.go.dev/github.com/septagon-oss/pk-testkit)
[![CI](https://github.com/septagon-oss/pk-testkit/actions/workflows/go.yml/badge.svg)](https://github.com/septagon-oss/pk-testkit/actions/workflows/go.yml)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Composable, adapter-neutral conformance contracts for the OSS PlatformKit family. `pk-testkit` lets module authors and app teams prove public claims, run API flows, and validate requirement-to-flow coverage without binding to a specific browser, database, container runtime, or hosted CI service. Downstream distributions can adapt the same contracts to Playwright, Kubernetes, SaaS smoke tests, or private E2E harnesses.

## Install

```bash
go get github.com/septagon-oss/pk-testkit@v0.1.0
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/septagon-oss/pk-testkit/pkg/conformance"
)

func main() {
	suite, err := conformance.NewSuite(conformance.Check{
		ID:            "runtime.ready",
		RequirementID: "REQ-READY",
		Description:   "runtime reports ready",
		Run: func(context.Context) (conformance.Result, error) {
			return conformance.Pass("runtime is ready"), nil
		},
	})
	if err != nil {
		panic(err)
	}

	report := suite.Run(context.Background())
	fmt.Println(report.Status, report.OK())
}
```

## Current Surface

- `pkg/conformance`: deterministic checks and reports for runtime/module conformance.
- `pkg/flowtest`: requirement-to-flow coverage validation using `pk-shared` flow definitions.
- `pkg/apitest`: standard-library API flow runner over an `http.Handler`.

## Verify

```bash
make verify   # go test + go vet + staticcheck + race
```

## License

Apache-2.0. See [LICENSE](LICENSE).
