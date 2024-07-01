package testkit

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/mumoshu/testkit/datadog"
)

// RequireDatadogMetrics queries metrics from Datadog and validates them.
//
// query can be a query string like the following examples:
// -  "avg:kubernetes.cpu.user.total{*}"
// -  "system.cpu.idle{*}"
//
// validate is called for each series in the response.
// The test fails if any of the validate calls return an error.
// The test fails also if the query returns no series.
func RequireDatadogMetrics(t *testing.T, query string, validate func(datadogV1.MetricsQueryMetadata) error) {
	t.Helper()

	if err := datadog.QueryAndValidateMetrics(query, validate); err != nil {
		t.Fatal(err)
	}
}
