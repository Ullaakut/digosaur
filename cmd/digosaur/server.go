package main

import (
	"context"
	"fmt"

	"github.com/Ullaakut/digosaur/api"
	"github.com/Ullaakut/digosaur/pkg/loki"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/hamba/pkg/v2/http/server"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func runServer(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	obsvr, err := observe.NewFromCLI(c, svcName, &observe.Options{
		LogTimestamps: true,
		LogTimeFormat: logger.TimeFormatISO8601,
		TracingAttrs:  []attribute.KeyValue{semconv.ServiceVersionKey.String(version)},
		StatsRuntime:  true,
	})
	if err != nil {
		return fmt.Errorf("creating observer: %w", err)
	}
	defer obsvr.Close()

	loki, err := loki.New(c.String(flagLokiAddr), obsvr)
	if err != nil {
		return fmt.Errorf("creating loki client: %w", err)
	}

	addr := c.String(flagAddr)
	srv := server.GenericServer[context.Context]{
		Addr:  addr,
		Stats: obsvr.Stats,
		Log:   obsvr.Log,
	}

	srv.Handler = api.New(loki, obsvr)

	obsvr.Log.Info("Starting server", lctx.Str("addr", addr))

	if err = srv.Run(ctx); err != nil {
		obsvr.Log.Error("Server error", lctx.Err(err))

		cancel()
	}

	return nil
}
