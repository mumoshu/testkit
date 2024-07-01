package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// QueryAndValidateMetrics queries metrics from Datadog.
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
// The validate function is called for each series in the response.
// You can use this function to validate the metadata of the series.
// The whole QueryAndValidateMetrics function returns an error if any of the validate calls return an error.
func QueryAndValidateMetrics(query string, validate func(datadogV1.MetricsQueryMetadata) error) error {
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewMetricsApi(apiClient)
	resp, r, err := api.QueryMetrics(ctx, time.Now().AddDate(0, 0, -1).Unix(), time.Now().Unix(), query)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MetricsApi.QueryMetrics`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return fmt.Errorf("error when calling `MetricsApi.QueryMetrics`: %v", err)
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprintf(os.Stdout, "Response from `MetricsApi.QueryMetrics`:\n%s\n", responseContent)

	if len(resp.GetSeries()) == 0 {
		return fmt.Errorf("no series found in the response")
	}

	for i, metadata := range resp.GetSeries() {
		if err := validate(metadata); err != nil {
			return fmt.Errorf("error when calling metadataCallback for series %d: %v", i, err)
		}
	}

	return nil
}
