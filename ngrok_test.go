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

func TestNgrok(t *testing.T) {
	conf := NgrokConfigFromEnv()
	ln, err := ListenNgrok(t, conf)
	require.NoError(t, err)

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

	res, err := http.Get(conf.EndpointURL)
	require.NoError(t, err)

	bodyData, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	body := string(bodyData)
	require.Equal(t, "Hello from ngrok-go!\n", body)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from ngrok-go!")
}
