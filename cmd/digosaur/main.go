package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ettle/strcase"
	"github.com/hamba/cmd/v2"
	"github.com/urfave/cli/v2"
)

const (
	flagAddr     = "addr"
	flagLokiAddr = "loki.addr"
)

const svcName = "digosaur"

var version = "¯\\_(ツ)_/¯"

var flags = cmd.Flags{
	&cli.StringFlag{
		Name:    flagAddr,
		Usage:   "The address to listen on",
		Value:   ":8080",
		EnvVars: []string{strcase.ToSNAKE(flagAddr)},
	},
	&cli.StringFlag{
		Name:    flagLokiAddr,
		Usage:   "The address of the Loki endpoint to send health data to",
		Value:   "http://localhost:3100",
		EnvVars: []string{strcase.ToSNAKE(flagLokiAddr)},
	},
}.Merge(cmd.MonitoringFlags)

func main() {
	os.Exit(realMain(os.Args))
}

func realMain(args []string) (code int) {
	defer func() {
		if v := recover(); v != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Panic: %v\n%s\n", v, debug.Stack())
			code = 1
		}
	}()
	app := cli.NewApp()
	app.Name = "Digosaur"
	app.Description = "Health dashboard backend"
	app.Version = version
	app.Flags = flags
	app.Action = runServer
	app.Suggest = true

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.RunContext(ctx, args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return 1
	}
	return 0
}
