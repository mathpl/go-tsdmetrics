package prometheusreporter

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	tsdmetrics "github.com/mathpl/go-tsdmetrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
)

type PrometheusConfig struct {
	namespace     string
	Registry      tsdmetrics.TaggedRegistry // Registry to be exported
	subsystem     string
	promRegistry  prometheus.Registerer //Prometheus registry
	FlushInterval time.Duration         //interval to update prom metrics
	gauges        map[string]prometheus.Gauge
}

// NewPrometheusProvider returns a Provider that produces Prometheus metrics.
// Namespace and subsystem are applied to all produced metrics.
func NewPrometheusProvider(r tsdmetrics.TaggedRegistry, namespace string, subsystem string, promRegistry prometheus.Registerer, FlushInterval time.Duration) *PrometheusConfig {
	return &PrometheusConfig{
		namespace:     namespace,
		subsystem:     subsystem,
		Registry:      r,
		promRegistry:  promRegistry,
		FlushInterval: FlushInterval,
		gauges:        make(map[string]prometheus.Gauge),
	}
}

func (c *PrometheusConfig) flattenKey(key string) string {
	key = strings.Replace(key, " ", "_", -1)
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	key = strings.Replace(key, "=", "_", -1)
	return key
}

func tagsToString(t tsdmetrics.Tags) string {
	var buf bytes.Buffer

	buf.Write([]byte("{"))

	l := len(t)
	cnt := 0
	for k, v := range t {
		if l != cnt {
			buf.Write([]byte(fmt.Sprintf("%s=%s,", k, v)))
		} else {
			buf.Write([]byte(fmt.Sprintf("%s=%s", k, v)))
		}
		cnt++
	}
	buf.Write([]byte("}"))

	return buf.String()
}

func (c *PrometheusConfig) gaugeFromNameAndValue(name string, t tsdmetrics.Tags, val float64) {
	key := fmt.Sprintf("%s_%s_%s%s", c.namespace, c.subsystem, name, tagsToString(t))
	g, ok := c.gauges[key]
	if !ok {
		g = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: c.flattenKey(c.namespace),
			Subsystem: c.flattenKey(c.subsystem),
			Name:      c.flattenKey(name),
			Help:      name,
		})
		c.promRegistry.MustRegister(g)
		c.gauges[key] = g
	}
	g.Set(val)
}

// TaggedOpenTSDBWithConfig is a blocking exporter function just like TaggedOpenTSDB,
// but it takes a TaggedOpenTSDBConfig instead.
func (t *PrometheusConfig) Run(ctx context.Context) {
	tick := time.Tick(t.FlushInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			if err := t.UpdatePrometheusMetricsOnce(); nil != err {
				log.Error(err)
			}
		}
	}
}

func (t *PrometheusConfig) RunWithPreprocessing(ctx context.Context, fn []func(tsdmetrics.TaggedRegistry)) {
	tick := time.Tick(t.FlushInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			fmt.Println("boop")
			for _, f := range fn {
				f(t.Registry)
			}
			if err := t.UpdatePrometheusMetricsOnce(); nil != err {
				log.Println(err)
			}
		}
	}
}

func (c *PrometheusConfig) UpdatePrometheusMetricsOnce() error {
	c.Registry.Each(func(name string, tm tsdmetrics.TaggedMetric) {
		switch metric := tm.(type) {
		case metrics.Counter:
			c.gaugeFromNameAndValue(name, tm.GetTags(), float64(metric.Count()))
		case metrics.Gauge:
		case metrics.GaugeFloat64:
			c.gaugeFromNameAndValue(name, tm.GetTags(), float64(metric.Value()))
		case metrics.Histogram:
			samples := metric.Snapshot().Sample().Values()
			if len(samples) > 0 {
				lastSample := samples[len(samples)-1]
				c.gaugeFromNameAndValue(name, tm.GetTags(), float64(lastSample))
			}
		case tsdmetrics.IntegerHistogram:
			samples := metric.Snapshot().Sample().Values()
			if len(samples) > 0 {
				lastSample := samples[len(samples)-1]
				c.gaugeFromNameAndValue(name, tm.GetTags(), float64(lastSample))
			}
		case metrics.Meter:
		case metrics.Timer:
			lastSample := metric.Snapshot().Rate1()
			c.gaugeFromNameAndValue(name, tm.GetTags(), float64(lastSample))
		}
	})
	return nil
}
