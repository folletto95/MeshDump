package meshdump

import (
	"encoding/json"
	"fmt"
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

	var data []Telemetry
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
