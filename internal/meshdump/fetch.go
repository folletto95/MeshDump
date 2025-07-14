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
// This is a minimal placeholder implementation that expects the node to
// expose JSON telemetry at /api/telemetry.
func FetchTelemetry(host string) ([]Telemetry, error) {
	url := fmt.Sprintf("http://%s/api/telemetry", host)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var data []Telemetry
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode telemetry: %v: %s", err, bytes.TrimSpace(body))
	}
	return data, nil
}
