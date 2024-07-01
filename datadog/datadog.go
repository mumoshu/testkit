package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// RequireMetrics queries metrics from Datadog and validates them.
//
// query can be a query string like the following examples:
// -  "avg:kubernetes.cpu.user.total{*}"
// -  "system.cpu.idle{*}"
//
// Note that this function does not validate the query on its own.
// For example, the following query is invalid:
//
//	"avg:kubernetes.cpu.user.total"
//
// You get:
//
//	Error parsing query: \nunable to parse avg:kubernetes.cpu.user.total: Rule 'scope_expr' didn't match at ” (line 1, column 30).
//
// because it lacks the scope_expr, which is required for the query to be valid.
// This specific query should be:
//
//	"avg:kubernetes.cpu.user.total{*}"
//
// The test fails if the query returns no series, or if any of the validate calls return an error.
func RequireMetrics(t *testing.T, query string, validate func(datadogV1.MetricsQueryMetadata) error) {
	t.Helper()

	if err := QueryAndValidateMetrics(query, validate); err != nil {
		t.Fatal(err)
	}
}
