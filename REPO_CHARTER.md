# pk-testkit Charter

## Purpose

Conformance, flow test, and API test contracts for PlatformKit OSS. Shared test utilities that ensure cross-repo compatibility and contract adherence.

## In Scope

- Conformance suites (`pkg/conformance`): module, entity, and store conformance contracts
- Flow test runner (`pkg/flowtest`): state machine–based workflow coverage
- API test helpers (`pkg/apitest`): handler-level request/response execution

## Out of Scope

- Browser automation or E2E tests
- Performance or load testing tooling
- Test data generators or fixtures
- CI/CD workflow definitions

## Dependencies

- `github.com/septagon-oss/pk-shared` — flow definitions and composition descriptors
