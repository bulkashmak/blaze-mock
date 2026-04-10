# Key Design Decisions

| Decision              | Choice                                  | Rationale                                                         |
| --------------------- | --------------------------------------- | ----------------------------------------------------------------- |
| Stub definition style | Builder pattern                         | Reads naturally, IDE autocompletion, mirrors k6's imperative feel |
| Dynamic responses     | `BodyFunc` via `WillRespondWith`        | Core differentiator. Full Go language for response logic          |
| Matching order        | Insertion order (first registered wins) | Simple, predictable, debuggable                                   |
| Path params           | `{name}` syntax -> regex                | Familiar to Go developers (echo, chi patterns)                    |
| Thread safety         | `sync.RWMutex` in registry              | Concurrent reads during matching, safe writes                     |
| Persistence           | In-memory only (v1)                     | Mock servers are ephemeral. No database, no files                 |
| 404 diagnostics       | Request details + registered stubs      | Critical for debugging                                            |
| Package structure     | Single `blaze` package                  | Simple import path                                                |
| Templating            | None -- use `BodyFunc`                  | Full Go language > templating DSL                                 |
| Request-to-response   | Two options: Req() helper + Extract/Template | Req() for full Go power, Extract/Template for declarative cases |
