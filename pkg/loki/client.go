package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
)

// Log represents a log message entry from Loki.
type Log struct {
	Timestamp string `json:"ts"`
	Line      string `json:"line"`
}

// Stream represents a Loki stream.
type Stream struct {
	Labels map[string]string `json:"steam"`
	Values [][]string        `json:"values"`
}

// Payload is the full request payload for Loki.
type Payload struct {
	Streams []Stream `json:"streams"`
}

type Client struct {
	url *url.URL

	log *logger.Logger
}

// New returns a Loki Client.
func New(baseURL string, obsvr *observe.Observer) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing loki url: %w", err)
	}

	client := Client{
		url: u,
		log: obsvr.Log.With(lctx.Str("component", "loki")),
	}

	return &client, nil
}

// Send sends logs to a Loki server.
func (c *Client) Send(ctx context.Context, entries []Stream) error {
	// Serialize to JSON
	jsonData, err := json.Marshal(Payload{Streams: entries})
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	_ = os.WriteFile("log.json", jsonData, 0644)

	// Send HTTP request to Loki
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("preparing Loki request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("pushing logs to Loki: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected loki response code: %d", resp.StatusCode)
	}

	c.log.Info("Pushed logs to logi", lctx.Int("count", len(entries)), lctx.Int("status", resp.StatusCode))

	return nil
}
