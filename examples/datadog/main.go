package main

import (
	"flag"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/mumoshu/testkit/datadog"
)

// main is an example of how to use the datadog package.
// You need to set the DD_APP_KEY and DD_API_KEY environment variables
// to authenticate with the Datadog API.
//
// Usage:
//
// go run ./examples/datadog -query "avg:kubernetes.cpu.user.total{*}"
func main() {
	var q string

	flag.StringVar(&q, "query", "avg:kubernetes.cpu.user.total{*}", "Datadog query")
	flag.Parse()

	if err := datadog.QueryAndValidateMetrics(q, func(mqm datadogV1.MetricsQueryMetadata) error {
		return nil
	}); err != nil {
		panic(err)
	}
}
