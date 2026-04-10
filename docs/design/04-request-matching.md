# Request Matching Algorithm

When an HTTP request arrives at the mock handler:

1. Iterate stubs in insertion order.
2. For each stub, check all matcher fields. A stub matches only if **all** specified matchers pass. Unspecified fields are "match any".
3. Return the first fully-matching stub.
4. If no stub matches, return 404 with diagnostic info (incoming request details + all registered stubs).

## Path Parameters

`blaze.Get("/api/payments/{id}")` is internally converted to a regex `/api/payments/([^/]+)` with a named capture group. `blaze.PathParam(r, "id")` extracts the captured value from request context.
