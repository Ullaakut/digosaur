package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ullaakut/digosaur/api"
	"github.com/hamba/cmd/v2/observe"
)

var outPath = flag.String("o", "", "The path to write the openapi spec to")

func main() {
	flag.Parse()

	if *outPath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Output path is required")
		return
	}

	spec, err := serverSpec()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error: ", err)
		return
	}

	p := filepath.Join(*outPath, "server-spec.json")
	if err = os.WriteFile(p, spec, 0o600); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error: ", err)
		return
	}
}

func serverSpec() ([]byte, error) {
	obsrvFake := observe.NewFake()

	srv := api.New(nil, obsrvFake)

	specJSON, err := srv.OpenAPISpec()
	if err != nil {
		return nil, err
	}
	return specJSON, nil
}
