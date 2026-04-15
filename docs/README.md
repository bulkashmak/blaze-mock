# Blaze Mock Documentation

Blaze Mock is a lightweight, standalone HTTP mock server for QA testing. Users write Go scripts to configure stubs imperatively — the same way Grafana k6 lets users write load tests as imperative JavaScript rather than declarative YAML/JSON.

### Why not WireMock?

- WireMock requires JSON stub definitions (declarative). This makes conditional logic, loops, dynamic responses, and code reuse awkward.
- WireMock is a heavyweight Java process.
- Blaze lets you write Go code: full language power for matching logic and response construction, zero separate config files, easy to version-control.

## Documentation

1. [Core Concepts](core-concepts.md) — Stub, matchers, response definitions, request value extraction
2. [API Reference](api-reference.md) — Builder pattern, server API, matchers, extractors
3. [Architecture](architecture.md) — Component diagram, package layout, request matching
4. [Admin API](admin-api.md) - HTTP Admin API documentation
