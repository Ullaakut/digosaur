package influx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	inflx "github.com/influxdata/influxdb-client-go/v2"
)

// Client is an influx client.
type Client struct {
	inflx.Client

	org    string
	bucket string
}

// New creates a new Influx client.
func New(url, token, org, bucket string) *Client {
	client := inflx.NewClient(url, token)

	return &Client{
		Client: client,
		org:    org,
		bucket: bucket,
	}
}

// Point represents a metric point in time, and its metadata.
type Point struct {
	Name string    `json:"name"`
	Unit string    `json:"unit"`
	Date time.Time `date:"date"`

	Data Data `json:"data"`
}

// Data represents the actual data within a point.
type Data struct {
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
}

// Add adds a point to the Influx database.
func (c *Client) Add(point Point) error {
	m, err := toMap(point.Data)
	if err != nil {
		return fmt.Errorf("converting point data to map: %w", err)
	}

	writeAPI := c.WriteAPIBlocking(c.org, c.bucket)
	p := inflx.NewPoint(
		point.Name,
		map[string]string{"unit": point.Unit},
		m,
		point.Date,
	)

	return writeAPI.WritePoint(context.Background(), p)
}

func toMap(s any) (map[string]any, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
