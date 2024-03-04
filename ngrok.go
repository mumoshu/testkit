package testkit

import (
	"context"
	"os"
	"testing"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

type NgrokTunnelConfig struct {
	// EdgeLabel is the label of the ngrok edge to use.
	//
	// Let's say you want to establish an ngrok tunnel to your local HTTP server running on port 80,
	// you will usually run the following command to establish the tunnel:
	//   ngrok tunnel --label edge=edghts_<some randomish id> http://localhost:80
	//
	// In this case, the edge label is "edghts_<some randomish id>".
	EdgeLabel string

	// The URL of the ngrok tunnel endpoint.
	// This needs to be the one that corresponds to the edge associated to the edge label.
	EndpointURL string

	// Token is the ngrok authtoken.
	Token string
}

func NgrokConfigFromEnv() NgrokTunnelConfig {
	return NgrokTunnelConfig{
		EndpointURL: os.Getenv("TESTKIT_NGROK_ENDPOINT_URL"),
		EdgeLabel:   os.Getenv("TESTKIT_NGROK_EDGE_LABEL"),
		Token:       os.Getenv("TESTKIT_NGROK_AUTHTOKEN"),
	}
}

// ListenNgrok establishes an ngrok tunnel and listens for incoming HTTP requests.
// The caller is responsible for closing the returned listener.
func ListenNgrok(t *testing.T, c NgrokTunnelConfig) (ngrok.Tunnel, error) {
	if c.Token == "" {
		t.Fatal("ngrok token is not set")
	}

	if c.EdgeLabel == "" {
		t.Fatal("ngrok edge label is not set")
	}

	ctx := context.Background()

	tunnel := config.LabeledTunnel(config.WithLabel("edge", c.EdgeLabel))

	ln, err := ngrok.Listen(ctx,
		tunnel,
		ngrok.WithAuthtoken(c.Token),
	)
	if err != nil {
		return nil, err
	}

	// Automatically close the listener when the test is done.
	t.Cleanup(func() {
		if err := ln.Close(); err != nil {
			t.Logf("ln.Close: %v", err)
		}
	})

	t.Logf("Ingress established at: %s", ln.URL())

	return ln, nil
}
