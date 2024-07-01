package testkit

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test and demonstrate how to use the Ngrok provider of TestKit
// to expose a local server to the internet and run tests against it.
//
// Use this as a reference to write your own tests against your server
// application using the Ngrok provider of TestKit.
func TestNgrok(t *testing.T) {
	// Start a Ngrok tunnel using TestKit

	conf := NgrokConfigFromEnv()
	ln, err := ListenNgrok(t, conf)
	require.NoError(t, err)

	// Start a server to be exposed via the Ngrok tunnel
	// This is where you would typically start your server being tested.

	go func() {
		if err := http.Serve(ln, http.HandlerFunc(handler)); err != nil {
			t.Logf("http.Serve: %v", err)
		}
	}()

	// Wait for the tunnel to be established
	// and the server to be ready to serve requests
	// before running the following assertion code.

	log.Printf("Ingress established at: %s", conf.EndpointURL)

	time.Sleep(10 * time.Second)

	// Verifies that the server is accessible via the ngrok tunnel
	// This is where you would typically run your tests against the server.

	res, err := http.Get(conf.EndpointURL)
	require.NoError(t, err)

	bodyData, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	body := string(bodyData)
	require.Equal(t, "Hello from ngrok-go!\n", body)
}

// handler is a simple HTTP handler that responds with a message.
// In a real-world scenario, this would be your server being tested.
// It can be an HTTP server, a gRPC server, or any other server.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from ngrok-go!")
}
