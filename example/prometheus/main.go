package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	tsdmetrics "github.com/mathpl/go-tsdmetrics"
	"github.com/mathpl/go-tsdmetrics/prometheusreporter"
)

func main() {
	rootRegistry := tsdmetrics.NewSegmentedTaggedRegistry("", nil, nil)
	tsdmetrics.RegisterTaggedRuntimeMemStats(rootRegistry)

	promReporter := prometheusreporter.NewPrometheusProvider(rootRegistry, "test", "test", prometheus.DefaultRegisterer, 1*time.Second)
	go promReporter.RunWithPreprocessing(context.Background(), tsdmetrics.RuntimeCaptureFn)

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
