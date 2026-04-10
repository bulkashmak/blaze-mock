# Blaze Mock - Software Design Document

## Overview

Blaze Mock is a lightweight, standalone HTTP mock server for QA testing. Users write Go scripts to configure stubs imperatively - the same way Grafana k6 lets users write load tests as imperative JavaScript rather than declarative YAML/JSON.

### Why not WireMock?

- WireMock requires JSON stub definitions (declarative). This makes conditional logic, loops, dynamic responses, and code reuse awkward.
- WireMock is a heavyweight Java process.
- Blaze lets you write Go code: full language power for matching logic and response construction, zero separate config files, easy to version-control.

## Table of Contents

1. [Core Concepts](01-core-concepts.md) - Stub, matchers, response definitions
2. [Go API Design](02-go-api.md) - Builder pattern, usage examples, server API
3. [Architecture](03-architecture.md) - Component diagram, package layout, stub registry
4. [Request Matching](04-request-matching.md) - Matching algorithm, path parameters
5. [Design Decisions](05-design-decisions.md) - Key choices and rationale
6. [Dependencies](06-dependencies.md) - v1 dependency list
7. [Implementation Phases](07-implementation-phases.md) - Build order
