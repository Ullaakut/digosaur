package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Ullaakut/digosaur/pkg/influx"
	"github.com/gamefabric/openapi"
	kin "github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	mdlw "github.com/hamba/pkg/v2/http/middleware"
	"github.com/hamba/pkg/v2/http/render"
	"github.com/hamba/statter/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

//go:generate oapi-gen -all

// Store represents something that can send entries to Loki.
type Store interface {
	Add(ctx context.Context, pt influx.Point) error
}

// Server serves web api requests.
type Server struct {
	h http.Handler

	db Store

	log    *logger.Logger
	tracer trace.Tracer
}

// New returns a server.
func New(db Store, obsvr *observe.Observer) *Server {
	srv := &Server{
		db:     db,
		log:    obsvr.Log,
		tracer: obsvr.Tracer("api"),
	}

	srv.h = srv.routes(obsvr.Stats, obsvr.TraceProv, obsvr.Log)

	return srv
}

func (s *Server) routes(stats *statter.Statter, tp trace.TracerProvider, log *logger.Logger) http.Handler {
	mux := chi.NewRouter()
	mux.Use(mdlw.Recovery(log))
	mux.Use(mdlw.Tracing("server", otelhttp.WithTracerProvider(tp), otelhttp.WithPropagators(&propagation.TraceContext{})))

	// Add default documentation.
	mux.Use(openapi.Op().
		Produces("application/json").
		Returns(http.StatusInternalServerError, "Internal Server Error", &render.APIError{}).
		Build(),
	)

	mux.With(mdlw.Stats("apple", stats)).Post("/apple", s.handleApple())

	mux.Get("/api/openapi/v3", s.handleOpenAPI())

	return mux
}

// ServeHTTP serves http requests.
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(rw, r)
}

// OpenAPISpec returns the OpenAPI v3 spec for the server in JSON.
func (s *Server) OpenAPISpec() ([]byte, error) {
	doc, err := openapi.BuildSpec(s.h.(*chi.Mux), openapi.SpecConfig{})
	if err != nil {
		return nil, fmt.Errorf("generating spec: %w", err)
	}
	doc.Info = &kin.Info{
		Title:   "Template API",
		Version: "1",
	}

	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding spec: %w", err)
	}
	return b, nil
}

func (s *Server) handleOpenAPI() http.HandlerFunc {
	var cache []byte
	return func(rw http.ResponseWriter, _ *http.Request) {
		if len(cache) == 0 {
			b, err := s.OpenAPISpec()
			if err != nil {
				s.log.Error("Could not serve OpenAPI spec", lctx.Err(err))

				http.Error(rw, "internal server error", http.StatusInternalServerError)
				return
			}

			cache = b
		}

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write(cache)
	}
}
