package requirehttp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mumoshu/testkit/requirehttp"
	"github.com/stretchr/testify/require"
)

// This verifies the response status code and the response body
// using the requirehttp.JSONResponse function.
// The target is a simple HTTP server that returns a JSON response,
// inspired by http://ip.jsontest.com/, which returns the client's IP address.
func TestRequireHTTPJsonResponse(t *testing.T) {
	var respData struct {
		IP string `json:"ip"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ip":"127.0.0.1"}`)
	}))

	req, err := http.NewRequest("GET", srv.URL, nil)
	require.NoError(t, err)

	requirehttp.JSONResponse(t, req, http.StatusOK, &respData, func() error {
		if respData.IP == "" {
			return fmt.Errorf("unexpected message: %v", respData)
		}
		return nil
	})
}
