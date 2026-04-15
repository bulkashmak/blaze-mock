package main

import (
	"log"
	"net/http"

	"github.com/bulkashmak/blaze-mock/blaze"
)

func main() {
	server := blaze.NewServer(
		blaze.WithPort(8080),
		blaze.WithLogOutput(blaze.LogBoth),
		blaze.WithLogFile("blaze.log"),
	)

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

	// JSON body matching — structural equality (ignores key order)
	server.Stub(
		blaze.Post("/api/invoices").
			WithBody(blaze.EqualToJSON(`{"amount": 500, "currency": "EUR"}`)).
			WillReturn(
				blaze.Response(201).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"id": "inv_001", "status": "created"}`),
			),
	)

	// JSON body matching — combine multiple field matchers with AllOf
	server.Stub(
		blaze.Post("/api/refunds").
			WithBody(blaze.AllOf(
				blaze.MatchesJSONPath("$.reason", blaze.Contains("defective")),
				blaze.MatchesJSONPath("$.order_id", blaze.Prefix("ord_")),
			)).
			WillReturn(
				blaze.Response(202).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"status": "refund_pending"}`),
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
					WithBodyTemplate(`{"greeting": "Hello, {{.name}}","email": "{{.email}}","format": "{{.format}}"}`),
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

	log.Fatal(server.Start())
}
