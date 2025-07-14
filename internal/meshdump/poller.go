package meshdump

import (
	"context"
	"log"
	"time"
)

// PollNodes periodically fetches telemetry from the given node IP addresses
// and stores the results using the provided Store. It runs until the context
// is cancelled.
func PollNodes(ctx context.Context, interval time.Duration, store *Store, nodes []string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, n := range nodes {
				log.Printf("poller: fetching telemetry from %s", n)
				data, err := FetchTelemetry(n)
				if err != nil {
					log.Printf("fetch telemetry from %s: %v", n, err)
					continue
				}
				log.Printf("poller: received %d entries from %s", len(data), n)
				for _, t := range data {
					store.Add(t)
				}
			}
		}
	}
}
