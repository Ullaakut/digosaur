package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Ullaakut/digosaur/pkg/loki"
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

		var entries []loki.Stream
		for _, metric := range health.Data.Metrics {

			// Sort metrics in chronological order.
			slices.SortFunc(metric.Data, func(i, j Point) int {
				tsI, err := time.Parse("2006-01-02 15:04:05 -0700", i.Date)
				if err != nil {
					panic(err)
				}

				tsJ, err := time.Parse("2006-01-02 15:04:05 -0700", j.Date)
				if err != nil {
					panic(err)
				}

				if tsI.Before(tsJ) {
					return -1
				} else if tsI.After(tsJ) {
					return 1
				}
				return 0
			})

			values := make([][]string, 0)
			for _, point := range metric.Data {
				ts, err := time.Parse("2006-01-02 15:04:05 -0700", point.Date)
				if err != nil {
					s.log.Error("Failed to parse Apple Health data timestamp", lctx.Err(err))
					render.JSONError(rw, http.StatusBadRequest, fmt.Sprintf("invalid timestamp: %v", err))
					return
				}

				entry, err := line(metric.Name, metric.Units, point)
				if err != nil {
					s.log.Error("Failed to create Apple Health data log", lctx.Err(err))
					render.JSONError(rw, http.StatusBadRequest, fmt.Sprintf("invalid data: %v", err))
					return
				}

				values = append(values, []string{
					strconv.FormatInt(ts.UnixNano(), 10),
					string(entry),
				})
			}

			entries = append(entries, loki.Stream{
				Labels: map[string]string{
					"metric": metric.Name,
					"units":  metric.Units,
				},
				Values: values,
			})
		}

		s.log.Info("Received Apple Health data", lctx.Int("entries", len(entries)))

		err = s.sink.Send(req.Context(), entries)
		if err != nil {
			s.log.Error("Failed to send Apple Health data to Loki", lctx.Err(err))

			render.JSONInternalServerError(rw)
			return
		}

		rw.WriteHeader(http.StatusOK)
	})
}

func line(name, units string, point Point) ([]byte, error) {
	lineObj := struct {
		Name  string `json:"name"`
		Units string `json:"units"`

		Quantity   float64 `json:"qty,omitempty"`
		Average    float64 `json:"avg,omitempty"`
		Minimum    float64 `json:"min,omitempty"`
		Maximum    float64 `json:"max,omitempty"`
		Deep       float64 `json:"deep,omitempty"`
		Core       float64 `json:"core,omitempty"`
		Awake      float64 `json:"awake,omitempty"`
		Asleep     float64 `json:"asleep,omitempty"`
		SleepEnd   string  `json:"sleep_end,omitempty"`
		InBedStart string  `json:"in_bed_start,omitempty"`
		InBedEnd   string  `json:"in_bed_end,omitempty"`
		SleepStart string  `json:"sleep_start,omitempty"`
		Rem        float64 `json:"rem,omitempty"`
		InBed      float64 `json:"in_bed,omitempty"`
	}{
		Name:  name,
		Units: units,

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
	}

	return json.Marshal(lineObj)
}
