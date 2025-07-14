package meshdump

import (
        "context"
        "encoding/json"
        "log"

        mqtt "github.com/eclipse/paho.mqtt.golang"

)

// StartMQTT connects to the given broker and subscribes to the provided topic.
// If user is non-empty, the client authenticates with the provided username and password.
// Messages are expected to contain a JSON encoded Telemetry struct. Received
// telemetry is stored in the provided Store until the context is cancelled.
func StartMQTT(ctx context.Context, broker, topic, user, pass string, store *Store) error {
	log.Printf("mqtt: connecting to %s", broker)
	opts := mqtt.NewClientOptions().AddBroker(broker)
	if user != "" {
		opts.SetUsername(user).SetPassword(pass)
	}
	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		return t.Error()
	}
	log.Printf("mqtt: connected, subscribing to %s", topic)

	// send a welcome message to verify connectivity
	client.Publish("meshdump/welcome", 0, false, []byte("MeshDump connected"))

	if t := client.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
		var tel Telemetry
		if err := json.Unmarshal(m.Payload(), &tel); err != nil {
			log.Printf("mqtt decode: %v", err)
			return
		}
		log.Printf("mqtt: message from %s type=%s value=%f", tel.NodeID, tel.DataType, tel.Value)
		store.Add(tel)
	}); t.Wait() && t.Error() != nil {
		client.Disconnect(250)
		return t.Error()
	}
	log.Printf("mqtt: subscribed to %s", topic)

	go func() {
		<-ctx.Done()
		client.Disconnect(250)
	}()
	return nil
}
