# Implementation Phases

## Phase 1 - Core library

1. `matcher.go` - StringMatcher, BodyMatcher interfaces and implementations
2. `stub.go` - Stub struct, StubBuilder
3. `response.go` - ResponseDef, ResponseBuilder
4. `registry.go` - StubRegistry with Add, Remove, Reset, List, Match
5. `handler.go` - Mock HTTP handler
6. `server.go` - Server tying it all together
7. `options.go` - ServerOption functional options

## Phase 2 -- Polish

8. README with usage examples
9. Integration tests
