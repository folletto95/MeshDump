package meshdump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchTelemetry retrieves telemetry data from a remote Meshtastic node.
// According to the Meshtastic HTTP API documentation, nodes expose
// telemetry at the `/api/v1/telemetry` endpoint.
func FetchTelemetry(host string) ([]Telemetry, error) {
	url := fmt.Sprintf("http://%s/api/v1/telemetry", host)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var data []Telemetry
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode telemetry: %v: %s", err, bytes.TrimSpace(body))
	}
	return data, nil
}
