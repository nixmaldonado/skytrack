package opensky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://opensky-network.org/api"

var ErrRateLimited = errors.New("opensky: rate limited (HTTP 429)")

// Client fetches aircraft state vectors from the OpenSky Network REST API.
type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
	}
}

// GetAircraftState fetches the current state of a single aircraft by its ICAO24
// hex transponder address. Returns nil, nil if the aircraft is not broadcasting.
func (c *Client) GetAircraftState(ctx context.Context, icao24 string) (*StateVector, error) {
	url := fmt.Sprintf("%s/states/all?icao24=%s", c.baseURL, icao24)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("opensky: create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opensky: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("opensky: unexpected status %d", resp.StatusCode)
	}

	var apiResp struct {
		Time   int               `json:"time"`
		States []json.RawMessage `json:"states"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("opensky: decode response: %w", err)
	}

	if len(apiResp.States) == 0 {
		return nil, nil // aircraft not broadcasting
	}

	return parseStateVector(apiResp.States[0])
}

// parseStateVector parses a single OpenSky state array into a StateVector.
// OpenSky returns each state as a JSON array with positional indices:
//
//	[0] icao24, [1] callsign, [5] longitude, [6] latitude, [7] baro_altitude,
//	[8] on_ground, [9] velocity, [10] true_track, [11] vertical_rate
func parseStateVector(raw json.RawMessage) (*StateVector, error) {
	var fields []any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("opensky: unmarshal state: %w", err)
	}

	if len(fields) < 12 {
		return nil, fmt.Errorf("opensky: expected at least 12 fields, got %d", len(fields))
	}

	sv := &StateVector{}

	if v, ok := fields[0].(string); ok {
		sv.Icao24 = v
	}
	if v, ok := fields[1].(string); ok {
		sv.Callsign = strings.TrimSpace(v)
	}
	sv.Longitude = toFloat64Ptr(fields[5])
	sv.Latitude = toFloat64Ptr(fields[6])
	sv.BaroAltitude = toFloat64Ptr(fields[7])
	if v, ok := fields[8].(bool); ok {
		sv.OnGround = v
	}
	sv.Velocity = toFloat64Ptr(fields[9])
	sv.TrueTrack = toFloat64Ptr(fields[10])
	sv.VerticalRate = toFloat64Ptr(fields[11])

	if v := toFloat64Ptr(fields[4]); v != nil {
		ts := int64(*v)
		sv.TimePosition = &ts
	}

	return sv, nil
}

func toFloat64Ptr(v any) *float64 {
	if v == nil {
		return nil
	}
	if f, ok := v.(float64); ok {
		return &f
	}
	return nil
}
