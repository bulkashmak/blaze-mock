# Blaze Mock - Admin API Sample

Demonstrates the HTTP admin API for runtime stub management.

## What this shows

- Starting a server with both mock (`:8080`) and admin (`:8081`) ports
- Seeding a Go code stub visible via the admin API
- Full CRUD lifecycle: create, read, update, delete stubs via HTTP
- Deleting all stubs at once

## Running

From the repository root:

```bash
go run ./samples/admin-api/
```

## Testing

Run the provided curl scripts against the running server:

```bash
./samples/admin-api/requests.sh
```
