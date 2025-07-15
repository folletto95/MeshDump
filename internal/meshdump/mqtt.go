package meshdump

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	meshtastic "github.com/meshtastic/go/generated"
	"google.golang.org/protobuf/proto"
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

	debug := store.debug
	if t := client.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
		if debug {
			log.Printf("TOPIC: %s", m.Topic())
			log.Printf("PAYLOAD HEX: %s", hex.EncodeToString(m.Payload()))
			log.Printf("PAYLOAD STRING: %s", string(m.Payload()))
		}

		payload := m.Payload()
		if len(payload) == 0 {
			return
		}

		if payload[0] == '{' {
			var tel Telemetry
			if err := json.Unmarshal(payload, &tel); err != nil {
				log.Printf("mqtt decode json: %v", err)
				return
			}
			store.Add(tel)
			return
		}

		var pb meshtastic.Telemetry
		if err := proto.Unmarshal(payload, &pb); err != nil {
			log.Printf("mqtt decode protobuf: %v", err)
			return
		}

		nodeID, dataType := parseTopic(m.Topic())
		for _, t := range convertTelemetryProto(nodeID, dataType, &pb) {
			store.Add(t)
		}
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

// parseTopic attempts to derive node ID and data type from the MQTT topic.
// It returns the second to last and last path components when possible.
func parseTopic(topic string) (string, string) {
	parts := strings.Split(topic, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return "", ""
}

// convertTelemetryProto converts a Telemetry protobuf message into one or more
// Telemetry records understood by the store.
func convertTelemetryProto(nodeID, defaultType string, msg *meshtastic.Telemetry) []Telemetry {
	ts := time.Unix(int64(msg.GetTime()), 0)
	var out []Telemetry
	switch v := msg.Variant.(type) {
	case *meshtastic.Telemetry_DeviceMetrics:
		dm := v.DeviceMetrics
		if dm.BatteryLevel != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "battery_level", Value: float64(dm.GetBatteryLevel()), Timestamp: ts})
		}
		if dm.Voltage != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "voltage", Value: float64(dm.GetVoltage()), Timestamp: ts})
		}
		if dm.ChannelUtilization != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "channel_utilization", Value: float64(dm.GetChannelUtilization()), Timestamp: ts})
		}
		if dm.AirUtilTx != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "air_util_tx", Value: float64(dm.GetAirUtilTx()), Timestamp: ts})
		}
		if dm.UptimeSeconds != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "uptime_seconds", Value: float64(dm.GetUptimeSeconds()), Timestamp: ts})
		}
	case *meshtastic.Telemetry_EnvironmentMetrics:
		em := v.EnvironmentMetrics
		if em.Temperature != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "temperature", Value: float64(em.GetTemperature()), Timestamp: ts})
		}
		if em.RelativeHumidity != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "relative_humidity", Value: float64(em.GetRelativeHumidity()), Timestamp: ts})
		}
		if em.BarometricPressure != nil {
			out = append(out, Telemetry{NodeID: nodeID, DataType: "barometric_pressure", Value: float64(em.GetBarometricPressure()), Timestamp: ts})
		}
	default:
		if defaultType != "" {
			// use the provided type with a zero value if we cannot decode
			out = append(out, Telemetry{NodeID: nodeID, DataType: defaultType, Value: 0, Timestamp: ts})
		}
	}
	return out
}
