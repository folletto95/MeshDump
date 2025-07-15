package meshdump

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"
	pproto "meshdump/internal/proto"
)

// nodeIDFromTopic attempts to extract a node ID from a MQTT topic. The default
// Meshtastic topic format is "msh/<nodeId>/..." so we return the first segment
// after the root if present.
func nodeIDFromTopic(topic string) (string, bool) {
	parts := strings.Split(topic, "/")
	if len(parts) >= 2 {
		return parts[1], true
	}
	if len(parts) == 1 {
		return parts[0], true
	}
	return "", false
}

// StartMQTT connects to the given broker and subscribes to the provided topic.
// If user is non-empty, the client authenticates with the provided username and password.
// Incoming messages are first decoded as JSON Telemetry. If that fails they are
// treated as protobuf MapReport messages. Decoded telemetry or node info is
// stored in the provided Store until the context is cancelled.
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
		if err := json.Unmarshal(m.Payload(), &tel); err == nil {
			if tel.NodeID == "" {
				if id, ok := nodeIDFromTopic(m.Topic()); ok {
					tel.NodeID = id
				}
			}
			if tel.NodeID == "" {
				log.Printf("mqtt: telemetry message missing node id: %s", m.Topic())
				return
			}
			log.Printf("mqtt: message from %s type=%s value=%f", tel.NodeID, tel.DataType, tel.Value)
			store.Add(tel)
			return
		}

		// not JSON telemetry, try protobuf map report
		var mr pproto.MapReport
		if err := proto.Unmarshal(m.Payload(), &mr); err == nil {
			if id, ok := nodeIDFromTopic(m.Topic()); ok {
				info := NodeInfo{ID: id, LongName: mr.GetLongName(), ShortName: mr.GetShortName(), Firmware: mr.GetFirmwareVersion()}
				store.SetNodeInfo(info)
				log.Printf("mqtt: map report for %s firmware=%s", id, mr.GetFirmwareVersion())
			} else {
				log.Printf("mqtt: map report received but topic missing node id: %s", m.Topic())
			}
			return
		}

		log.Printf("mqtt decode: unknown payload")

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
