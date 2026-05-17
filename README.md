# pk-testkit

Composable OSS conformance contracts for PlatformKit.

This repo proves requirements and flows without choosing a browser, database,
container runtime, or hosted CI service. Downstream distributions can adapt the
same contracts to Playwright, Kubernetes, SaaS smoke tests, or private E2E
harnesses.

The intent is simple: if PlatformKit makes a public claim, there should be a
portable way to prove it. `pk-testkit` gives module authors and app teams a
shared language for conformance reports, API flows, and requirement-to-flow
coverage.

## Current Surface

- `pkg/conformance`: deterministic checks and reports for runtime/module
  conformance.
- `pkg/flowtest`: requirement-to-flow coverage validation using `pk-shared`
  flow definitions.
- `pkg/apitest`: standard-library API flow runner over an `http.Handler`.

## Verify

```bash
make verify
make staticcheck
```
