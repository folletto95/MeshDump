package meshdump

import (
	"context"
	"encoding/json"
	"log"

	mqtt "meshdump/internal/stubmqtt"
)

// StartMQTT connects to the given broker and subscribes to the provided topic.
// If user is non-empty, the client authenticates with the provided username and password.
// Messages are expected to contain a JSON encoded Telemetry struct. Received
// telemetry is stored in the provided Store until the context is cancelled.
func StartMQTT(ctx context.Context, broker, topic, user, pass string, store *Store) error {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	if user != "" {
		opts.SetUsername(user).SetPassword(pass)
	}
	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		return t.Error()
	}

	if t := client.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
		var tel Telemetry
		if err := json.Unmarshal(m.Payload(), &tel); err != nil {
			log.Printf("mqtt decode: %v", err)
			return
		}
		store.Add(tel)
	}); t.Wait() && t.Error() != nil {
		client.Disconnect(250)
		return t.Error()
	}

	go func() {
		<-ctx.Done()
		client.Disconnect(250)
	}()
	return nil
}
