package types

import (
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"

	"github.com/tharsis/evmos/app"
)

func EnableObservability() error {
	if true {
		// Temporarily disabling this until we can configure out port reuse
		// fast enough or enabling observability through the config.
		// Please see https://github.com/tharsis/evmos/issues/84
		return nil
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "evmosd",
	})
	if err != nil {
		return fmt.Errorf("cmd/config: failed to create the OpenCensus Prometheus exporter: %w", err)
	}

	views := app.ObservabilityViews()
	if err := view.Register(views...); err != nil {
		return fmt.Errorf("cmd/config: failed to register OpenCensus views: %w", err)
	}
	view.RegisterExporter(pe)

	mux := http.NewServeMux()
	mux.Handle("/metrics", pe)

	// TODO: Derive the Prometheus observability exporter from the Evmos config.
	addr := ":8877"
	go func() {
		println("Serving the Prometheus observability exporter at", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			panic(err)
		}
	}()

	return nil
}
