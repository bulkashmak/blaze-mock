package main

import (
	"log"
	"net/http"

	"github.com/bulkashmak/blaze-mock/blaze"
)

func main() {
	server := blaze.NewServer(
		blaze.WithPort(8080),
		blaze.WithAdminPort(8081),
		blaze.WithLogOutput(blaze.LogStdout),
	)

	// Seed a stub via Go code — visible in admin API as a "code" stub
	server.Stub(
		blaze.Get("/api/health").
			WithID("health-check").
			WillRespondWith(func(r *http.Request) blaze.Resp {
				return blaze.Response(200).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"status": "ok", "source": "code"}`)
			}),
	)

	log.Printf("Mock server on :8080, Admin API on :8081")
	log.Fatal(server.Start())
}
