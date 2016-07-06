package tsdmetrics

import "fmt"

// DuplicateMetric is the error returned by Registry.Register when a metric
// already exists.  If you mean to Register that metric you must first
// Unregister the existing metric.
type DuplicateTaggedMetric struct {
	name string
	tags Tags
}

func (err DuplicateTaggedMetric) Error() string {
	return fmt.Sprintf("duplicate metric: %s %s", err.name, err.tags.String())
}
