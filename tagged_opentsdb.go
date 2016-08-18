package tsdmetrics

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/rcrowley/go-metrics"
)

type OpenTSDBFormat int

const (
	Tcollector OpenTSDBFormat = iota
	Json
)

type OpenTSDBPoint struct {
	Metric    string            `json:"metric"`
	Value     interface{}       `json:"value"`
	Timestamp int64             `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

// TaggedOpenTSDBConfig provides a container with configuration parameters for
// the TaggedOpenTSDB exporter
type TaggedOpenTSDB struct {
	Addr          string         // Network address to connect to
	Registry      TaggedRegistry // Registry to be exported
	FlushInterval time.Duration  // Flush interval
	DurationUnit  time.Duration  // Time conversion unit for durations
	Format        OpenTSDBFormat

	Logger *log.Logger

	netAddr *net.TCPAddr
}

// TaggedOpenTSDBWithConfig is a blocking exporter function just like TaggedOpenTSDB,
// but it takes a TaggedOpenTSDBConfig instead.
func (t *TaggedOpenTSDB) Run(ctx context.Context) {
	tick := time.Tick(t.FlushInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			if err := t.taggedOpenTSDB(); nil != err {
				t.Logger.Error(err)
			}
		}
	}
}

func (t *TaggedOpenTSDB) RunWithPreprocessing(ctx context.Context, fn []func(TaggedRegistry)) {
	tick := time.Tick(t.FlushInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick:
			for _, f := range fn {
				f(t.Registry)
			}
			if err := t.taggedOpenTSDB(); nil != err {
				t.Logger.Println(err)
			}
		}
	}
}

func (t *TaggedOpenTSDB) taggedOpenTSDB() error {
	now := time.Now().Unix()
	du := float64(t.DurationUnit)

	if t.Format == Tcollector {
		conn, err := net.DialTCP("tcp", nil, t.netAddr)
		if nil != err {
			return err
		}
		defer conn.Close()
		conn.SetWriteDeadline(time.Now().Add(time.Second * t.FlushInterval))

		w := bufio.NewWriter(conn)
		t.Registry.Each(func(name string, tm TaggedMetric) {
			tags := tm.GetTags()
			switch metric := tm.GetMetric().(type) {
			case metrics.Counter:
				fmt.Fprintf(w, "put %s %d %d %s\n", name, now, metric.Count(), tags.String())
			case metrics.Gauge:
				fmt.Fprintf(w, "put %s %d %d %s\n", name, now, metric.Value(), tags.String())
			case metrics.GaugeFloat64:
				fmt.Fprintf(w, "put %s %d %f %s\n", name, now, metric.Value(), tags.String())
			case metrics.Histogram:
				h := metric.Snapshot()
				ps := h.Percentiles([]float64{0.5, 0.75, 0.90, 0.95, 0.99})
				fmt.Fprintf(w, "put %s.count %d %d %s\n", name, now, h.Count(), tags.String())
				fmt.Fprintf(w, "put %s.min %d %d %s\n", name, now, h.Min(), tags.String())
				fmt.Fprintf(w, "put %s.max %d %d %s\n", name, now, h.Max(), tags.String())
				fmt.Fprintf(w, "put %s.mean %d %.2f %s\n", name, now, h.Mean(), tags.String())
				fmt.Fprintf(w, "put %s.std-dev %d %.2f %s\n", name, now, h.StdDev(), tags.String())
				fmt.Fprintf(w, "put %s.p50 %d %.2f %s\n", name, now, ps[0], tags.String())
				fmt.Fprintf(w, "put %s.p75 %d %.2f %s\n", name, now, ps[1], tags.String())
				fmt.Fprintf(w, "put %s.p90 %d %.2f %s\n", name, now, ps[2], tags.String())
				fmt.Fprintf(w, "put %s.p95 %d %.2f %s\n", name, now, ps[3], tags.String())
				fmt.Fprintf(w, "put %s.p99 %d %.2f %s\n", name, now, ps[4], tags.String())
			case metrics.Meter:
				m := metric.Snapshot()
				fmt.Fprintf(w, "put %s %d %d %s\n", name, now, m.Count(), tags.String())
				fmt.Fprintf(w, "put %s.1m-rate %d %.2f %s\n", name, now, m.Rate1(), tags.String())
				fmt.Fprintf(w, "put %s.5m-rate %d %.2f %s\n", name, now, m.Rate5(), tags.String())
				fmt.Fprintf(w, "put %s.15m-rate %d %.2f %s\n", name, now, m.Rate15(), tags.String())
				fmt.Fprintf(w, "put %s.mean-rate %d %.2f %s\n", name, now, m.RateMean(), tags.String())
			case metrics.Timer:
				t := metric.Snapshot()
				ps := t.Percentiles([]float64{0.5, 0.75, 0.90, 0.95, 0.99})
				fmt.Fprintf(w, "put %s.count %d %d %s\n", name, now, t.Count(), tags.String())
				fmt.Fprintf(w, "put %s.min %d %d %s\n", name, now, t.Min()/int64(du), tags.String())
				fmt.Fprintf(w, "put %s.max %d %d %s\n", name, now, t.Max()/int64(du), tags.String())
				fmt.Fprintf(w, "put %s.mean %d %.2f %s\n", name, now, t.Mean()/du, tags.String())
				fmt.Fprintf(w, "put %s.std-dev %d %.2f %s\n", name, now, t.StdDev()/du, tags.String())
				fmt.Fprintf(w, "put %s.p50 %d %.2f %s\n", name, now, ps[0]/du, tags.String())
				fmt.Fprintf(w, "put %s.p75 %d %.2f %s\n", name, now, ps[1]/du, tags.String())
				fmt.Fprintf(w, "put %s.p90 %d %.2f %s\n", name, now, ps[2]/du, tags.String())
				fmt.Fprintf(w, "put %s.p95 %d %.2f %s\n", name, now, ps[3]/du, tags.String())
				fmt.Fprintf(w, "put %s.p99 %d %.2f %s\n", name, now, ps[4]/du, tags.String())
				fmt.Fprintf(w, "put %s.1m-rate %d %.2f %s\n", name, now, t.Rate1(), tags.String())
				fmt.Fprintf(w, "put %s.5m-rate %d %.2f %s\n", name, now, t.Rate5(), tags.String())
				fmt.Fprintf(w, "put %s.15m-rate %d %.2f %s\n", name, now, t.Rate15(), tags.String())
				fmt.Fprintf(w, "put %s.mean-rate %d %.2f %s\n", name, now, t.RateMean(), tags.String())
			}
			w.Flush()
		})
	} else if t.Format == Json {
		var tsd []OpenTSDBPoint
		t.Registry.Each(func(name string, tm TaggedMetric) {
			tags := tm.GetTags()
			switch metric := tm.GetMetric().(type) {
			case metrics.Counter:
				tsd = append(tsd, OpenTSDBPoint{Metric: name, Timestamp: now, Value: metric.Count(), Tags: tags})
			case metrics.Gauge:
				tsd = append(tsd, OpenTSDBPoint{Metric: name, Timestamp: now, Value: metric.Value(), Tags: tags})
			case metrics.GaugeFloat64:
				tsd = append(tsd, OpenTSDBPoint{Metric: name, Timestamp: now, Value: metric.Value(), Tags: tags})
			case metrics.Histogram:
				h := metric.Snapshot()
				ps := h.Percentiles([]float64{0.5, 0.75, 0.90, 0.95, 0.99})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".count", Timestamp: now, Value: h.Count(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".min", Timestamp: now, Value: h.Min(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".max", Timestamp: now, Value: h.Max(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".mean", Timestamp: now, Value: h.Mean(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".std-dev", Timestamp: now, Value: h.StdDev(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p50", Timestamp: now, Value: ps[0], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p75", Timestamp: now, Value: ps[1], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p95", Timestamp: now, Value: ps[2], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p99", Timestamp: now, Value: ps[3], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p999", Timestamp: now, Value: ps[4], Tags: tags})
			case metrics.Meter:
				m := metric.Snapshot()
				tsd = append(tsd, OpenTSDBPoint{Metric: name, Timestamp: now, Value: m.Count(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".1m", Timestamp: now, Value: m.Rate1(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".5m", Timestamp: now, Value: m.Rate5(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".15m", Timestamp: now, Value: m.Rate15(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".mean-rate", Timestamp: now, Value: m.RateMean(), Tags: tags})
			case metrics.Timer:
				t := metric.Snapshot()
				ps := t.Percentiles([]float64{0.5, 0.75, 0.90, 0.95, 0.99})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".count", Timestamp: now, Value: t.Count(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".min", Timestamp: now, Value: t.Min(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".max", Timestamp: now, Value: t.Max(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".mean", Timestamp: now, Value: t.Mean(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".std-dev", Timestamp: now, Value: t.StdDev(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p50", Timestamp: now, Value: ps[0], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p75", Timestamp: now, Value: ps[1], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p95", Timestamp: now, Value: ps[2], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p99", Timestamp: now, Value: ps[3], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p999", Timestamp: now, Value: ps[4], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".1m", Timestamp: now, Value: t.Rate1(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".5m", Timestamp: now, Value: t.Rate5(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".15m", Timestamp: now, Value: t.Rate15(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".mean-rate", Timestamp: now, Value: t.RateMean(), Tags: tags})
			case IntegerHistogram:
				h := metric.Snapshot()
				ps := h.Percentiles([]float64{0.5, 0.75, 0.90, 0.95, 0.99})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".count", Timestamp: now, Value: h.Count(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".min", Timestamp: now, Value: h.Min(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".max", Timestamp: now, Value: h.Max(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".mean", Timestamp: now, Value: h.Mean(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".std-dev", Timestamp: now, Value: h.StdDev(), Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p50", Timestamp: now, Value: ps[0], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p75", Timestamp: now, Value: ps[1], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p95", Timestamp: now, Value: ps[2], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p99", Timestamp: now, Value: ps[3], Tags: tags})
				tsd = append(tsd, OpenTSDBPoint{Metric: name + ".p999", Timestamp: now, Value: ps[4], Tags: tags})
			}
		})

		if len(tsd) == 0 {
			t.Logger.Info("Nothing to send")
			return nil
		}

		if msg, err := json.Marshal(tsd); err != nil {
			t.Logger.Printf("Unable to serialize metrics json: %s", err)
		} else {
			contentReader := bytes.NewReader(msg)

			c := http.Client{Timeout: t.FlushInterval}
			if resp, err := c.Post(t.Addr, "application/json", contentReader); err != nil {
				t.Logger.Printf("Unable to send out metrics: %s", err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != 204 {
					t.Logger.Printf("Unexpected return code sending metrics: %d", resp.StatusCode)
				}
			}
		}
	}
	return nil
}
