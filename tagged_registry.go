package tsdmetrics

import (
	"reflect"
	"sync"

	"github.com/rcrowley/go-metrics"
)

// This is an interface so as to encourage other structs to implement
// the Registry API as appropriate.
type TaggedRegistry interface {
	// Call the given function for each registered metric.
	Each(func(string, TaggedMetric))
	WrappedEach(func(string, TaggedMetric) (string, TaggedMetric), func(string, TaggedMetric))

	// Get the metric by the given name or nil if none is registered.
	Get(string, Tags) interface{}

	// Gets an existing metric or registers the given one.
	// The interface can be the metric to register if not found in registry,
	// or a function returning the metric for lazy instantiation.
	GetOrRegister(string, Tags, interface{}) interface{}

	// Register the given metric under the given name.
	Register(string, Tags, interface{}) error

	// Run all registered healthchecks.
	RunHealthchecks()

	// Unregister the metric with the given name.
	Unregister(string, Tags)

	// Unregister all metrics.  (Mostly for testing.)
	UnregisterAll()
}

// The standard implementation of a Registry is a mutex-protected map
// of names to metrics.
type DefaultTaggedRegistry struct {
	metrics map[string]map[TagsID]TaggedMetric
	mutex   sync.Mutex
}

// Create a new registry.
func NewTaggedRegistry() TaggedRegistry {
	var r DefaultTaggedRegistry
	r.metrics = make(map[string]map[TagsID]TaggedMetric, 0)
	return &r
}

// Call the given function for each registered metric.
func (r *DefaultTaggedRegistry) Each(f func(string, TaggedMetric)) {
	for name, taggedMetrics := range r.registered() {
		for _, i := range taggedMetrics {
			f(name, i)
		}
	}
}

func (r *DefaultTaggedRegistry) WrappedEach(wrapperFunc func(string, TaggedMetric) (string, TaggedMetric), f func(string, TaggedMetric)) {
	for name, taggedMetrics := range r.registered() {
		for _, i := range taggedMetrics {
			f(wrapperFunc(name, i))
		}
	}
}

// Get the metric by the given name or nil if none is registered.
func (r *DefaultTaggedRegistry) Get(name string, tags Tags) interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if t, ok := r.metrics[name]; ok {
		if taggedMetric, ok := t[tags.TagsID()]; ok {
			return taggedMetric.GetMetric()
		}
	}
	return nil
}

// Gets an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
// The interface can be the metric to register if not found in registry,
// or a function returning the metric for lazy instantiation.
func (r *DefaultTaggedRegistry) GetOrRegister(name string, tags Tags, i interface{}) interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if t, ok := r.metrics[name]; ok {
		if taggedMetric, ok := t[tags.TagsID()]; ok {
			return taggedMetric.GetMetric()
		}
	}
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}
	r.register(name, tags, i)
	return i
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func (r *DefaultTaggedRegistry) Register(name string, tags Tags, i interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.register(name, tags, i)
}

// Run all registered healthchecks.
func (r *DefaultTaggedRegistry) RunHealthchecks() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, t := range r.metrics {
		for _, i := range t {
			if h, ok := i.GetMetric().(metrics.Healthcheck); ok {
				h.Check()
			}
		}
	}
}

// Unregister the metric with the given name.
func (r *DefaultTaggedRegistry) Unregister(name string, tags Tags) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if t, ok := r.metrics[name]; ok {
		if _, ok := t[tags.TagsID()]; ok {
			delete(t, tags.TagsID())
		}

		if len(t) == 0 {
			delete(r.metrics, name)
		}
	}
}

// Unregister all metrics.  (Mostly for testing.)
func (r *DefaultTaggedRegistry) UnregisterAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name, _ := range r.metrics {
		delete(r.metrics, name)
	}
}

func (r *DefaultTaggedRegistry) register(name string, tags Tags, i interface{}) error {
	if t, ok := r.metrics[name]; ok {
		if _, ok := t[tags.TagsID()]; ok {
			return DuplicateTaggedMetric{name, tags}
		}
	}
	switch i.(type) {
	case metrics.Counter, metrics.Gauge, metrics.GaugeFloat64, metrics.Healthcheck, metrics.Histogram, metrics.Meter, metrics.Timer:
		if _, ok := r.metrics[name]; !ok {
			r.metrics[name] = make(map[TagsID]TaggedMetric, 1)
		}
		taggedMetric := DefaultTaggedMetric{Tags: tags, Metric: i}
		r.metrics[name][tags.TagsID()] = &taggedMetric
	}
	return nil
}

func (r *DefaultTaggedRegistry) registered() map[string]map[TagsID]TaggedMetric {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	metrics := make(map[string]map[TagsID]TaggedMetric, len(r.metrics))
	for name, i := range r.metrics {
		metrics[name] = i
	}
	return metrics
}
