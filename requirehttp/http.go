package requirehttp

import (
	"encoding/json"
	"net/http"
	"testing"
)

// JSONResponse sends an HTTP request and validates the response.
//
// The test fails when any of the following conditions are met:
// - The response status code is not equal to statusCode.
// - The response body cannot be decoded into respData.
// - The validate function returns an error.
//
// Example:
//
//	var respData struct {
//	    Message string `json:"message"`
//	}
//
//	requirehttp.JSONResponse(t, req, http.StatusOK, &respData, func() error {
//	    if respData.Message != "hello" {
//	        return fmt.Errorf("unexpected message: %s", respData.Message)
//	    }
//	    return nil
//	})
func JSONResponse(t *testing.T, req *http.Request, statusCode int, respData interface{}, validate func() error) {
	t.Helper()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != statusCode {
		t.Fatalf("unexpected status code: expected %d, got %d", statusCode, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}

	if err := validate(); err != nil {
		t.Fatalf("unexpected response: %v", err)
	}
}
