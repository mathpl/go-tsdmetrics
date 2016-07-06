package tsdmetrics

import (
	"testing"

	"github.com/rcrowley/go-metrics"
)

func TestRegistry(t *testing.T) {
	r1 := NewSegmentedTaggedRegistry("prefix1", Tags{}, nil)
	r2 := NewSegmentedTaggedRegistry("prefix2", Tags{}, r1)
	r3 := NewSegmentedTaggedRegistry("prefix3", Tags{}, r2)

	myCnt := &metrics.StandardCounter{0}
	err := r3.Register("test1", Tags{"tag": "1"}, myCnt)
	if err != nil {
		t.Fatal(err)
	}

	c := r2.Get("test1", Tags{"tag": "1"})
	if c, ok := c.(metrics.StandardCounter); ok {

	}
}
