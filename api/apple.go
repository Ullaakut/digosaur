package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Ullaakut/digosaur/pkg/influx"
	"github.com/gamefabric/openapi"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/hamba/pkg/v2/http/render"
)

// AppleRequest represents an Apple Health data export.
type AppleRequest struct {
	Data Data `json:"data"`
}

// Data represents the data within an export.
type Data struct {
	Metrics []Metric `json:"metrics"`
}

// Metric represents a specific metric type.
type Metric struct {
	Name  string  `json:"name"`
	Units string  `json:"units"`
	Data  []Point `json:"data"`
}

// Point represents a specific data point.
type Point struct {
	Date string `json:"date"`

	// Source of the device that produced the data.
	//
	//openapi:optional
	Source string `json:"source"`

	Quantity   float64 `json:"qty"`
	Average    float64 `json:"avg"`
	Minimum    float64 `json:"min"`
	Maximum    float64 `json:"max"`
	Deep       float64 `json:"deep"`
	Core       float64 `json:"core"`
	Awake      float64 `json:"awake"`
	Asleep     float64 `json:"asleep"`
	SleepEnd   string  `json:"sleep_end"`
	InBedStart string  `json:"in_bed_start"`
	InBedEnd   string  `json:"in_bed_end"`
	SleepStart string  `json:"sleep_start"`
	Rem        float64 `json:"rem"`
	InBed      float64 `json:"in_bed"`
}

func (s *Server) handleApple() http.HandlerFunc {
	docs := openapi.Op().ID("apple").
		Doc("Processes Apple Health data exports").
		Produces("application/json").
		Returns(http.StatusOK, "OK", nil).
		Returns(http.StatusBadRequest, "Bad Request", &render.APIError{}).
		BuildHandler()

	return docs(func(rw http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.Start(req.Context(), "apple-export")
		defer span.End()

		var health AppleRequest
		err := json.NewDecoder(req.Body).Decode(&health)
		if err != nil {
			render.JSONError(rw, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
			return
		}

		for _, metric := range health.Data.Metrics {
			for _, point := range metric.Data {
				if isEmpty(point) {
					continue
				}

				ts, err := time.Parse("2006-01-02 15:04:05 -0700", point.Date)
				if err != nil {
					s.log.Error("Failed to parse Apple Health data timestamp", lctx.Err(err))
					render.JSONError(rw, http.StatusBadRequest, fmt.Sprintf("invalid timestamp: %v", err))
					return
				}

				err = s.db.Add(req.Context(), influx.Point{
					Name: metric.Name,
					Unit: metric.Units,
					Date: ts,
					Data: influx.Data{
						Quantity:   point.Quantity,
						Average:    point.Average,
						Minimum:    point.Minimum,
						Maximum:    point.Maximum,
						Deep:       point.Deep,
						Core:       point.Core,
						Awake:      point.Awake,
						Asleep:     point.Asleep,
						SleepEnd:   point.SleepEnd,
						InBedStart: point.InBedStart,
						InBedEnd:   point.InBedEnd,
						SleepStart: point.SleepStart,
						Rem:        point.Rem,
						InBed:      point.InBed,
					},
				})
				if err != nil {
					s.log.Error("Failed to push metric to db",
						lctx.Err(err),
						lctx.Str("metric", metric.Name),
						lctx.Interface("point", point),
					)
					render.JSONInternalServerError(rw)
					return
				}
			}
		}

		s.log.Info("Received Apple Health data", lctx.Int("metrics", len(health.Data.Metrics)))

		rw.WriteHeader(http.StatusOK)
	})
}

func isEmpty(p Point) bool {
	return p.Quantity == 0 &&
		p.Average == 0 &&
		p.Minimum == 0 &&
		p.Maximum == 0 &&
		p.Deep == 0 &&
		p.Core == 0 &&
		p.Awake == 0 &&
		p.Asleep == 0 &&
		p.Rem == 0 &&
		p.InBed == 0 &&
		p.SleepEnd == "" &&
		p.InBedStart == "" &&
		p.InBedEnd == "" &&
		p.SleepStart == ""
}
