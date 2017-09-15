package tsdmetrics

import (
	"time"

	"context"
)

func ExampleTaggedOpenTSDB() {
	go TaggedOpenTSDB(context.Background(), DefaultTaggedRegistry, 1*time.Second, "some.prefix", ":2003", Tcollector)
}

func ExampleTaggedOpenTSDBWithConfig() {
	go TaggedOpenTSDBWithConfig(context.Background(), TaggedOpenTSDBConfig{
		Addr:          ":2003",
		Registry:      DefaultTaggedRegistry,
		FlushInterval: 1 * time.Second,
		DurationUnit:  time.Millisecond,
	})
}
