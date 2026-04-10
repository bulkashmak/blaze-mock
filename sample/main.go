package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bulkashmak/blaze-mock/blaze"
)

func main() {
	server := blaze.NewServer(blaze.WithPort(8080))

	// Static inline response
	server.Stub(
		blaze.Post("/api/payments").
			WithHeader("Content-Type", blaze.EqualTo("application/json")).
			WithBodyContaining(`"amount"`).
			WillReturn(
				blaze.Response(201).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"id": "pay_123", "status": "created"}`),
			),
	)

	// Response from a static JSON fixture file
	server.Stub(
		blaze.Get("/api/users").
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyFile("fixtures/users.json"),
			),
	)

	// --- Option A: Req() helper inside WillRespondWith ---
	// Full Go power for request-to-response mapping
	server.Stub(
		blaze.Post("/api/orders/{id}/confirm").
			WillRespondWith(func(r *http.Request) blaze.Resp {
				req := blaze.Req(r)
				return blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyJSON(map[string]any{
						"order_id": req.PathParam("id"),
						"customer": req.JSONPath("$.customer.name"),
						"email":    req.JSONPath("$.customer.email"),
						"source":   req.Header("X-Source"),
						"status":   "confirmed",
					})
			}),
	)

	// --- Option B: Extract + Template ---
	// Declarative extraction with template-based response
	server.Stub(
		blaze.Post("/api/echo").
			Extract("name", blaze.FromJSONPath("$.user.name")).
			Extract("email", blaze.FromJSONPath("$.user.email")).
			Extract("token", blaze.FromHeader("Authorization")).
			Extract("format", blaze.FromQueryParam("format")).
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithHeader("X-Auth", "{{.token}}").
					WithBodyTemplate(
						`{
							"greeting": "Hello, {{.name}}",
							"email": "{{.email}}",
							"format": "{{.format}}"
						}`,
					),
			),
	)

	// Extract + Template with path parameters
	server.Stub(
		blaze.Get("/api/users/{id}").
			Extract("user_id", blaze.FromPathParam("id")).
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBodyTemplate(`{"id": "{{.user_id}}", "name": "User {{.user_id}}"}`),
			),
	)

	// Simple health check
	server.Stub(
		blaze.Get("/health").
			WillReturn(
				blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"status": "ok"}`),
			),
	)

	fmt.Println("Blaze Mock server starting on :8080")
	fmt.Println()
	fmt.Println("Try these endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  GET  /api/users")
	fmt.Println("  GET  /api/users/42")
	fmt.Println("  POST /api/payments           (body with \"amount\")")
	fmt.Println("  POST /api/orders/123/confirm (Option A: Req() helper)")
	fmt.Println("  POST /api/echo?format=json   (Option B: Extract + Template)")

	log.Fatal(server.Start())
}
