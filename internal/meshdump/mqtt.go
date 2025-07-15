package meshdump

import (
	"context"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mpb "github.com/meshtastic/go/generated"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// jsonPosition models the JSON payload sent by Meshtastic nodes when reporting
// their position. Coordinates are expressed as integers with seven decimal
// digits of precision. The struct only includes fields we care about.
type jsonPosition struct {
	Type      string `json:"type"`
	From      uint32 `json:"from"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Payload   struct {
		LatitudeI  int32 `json:"latitude_i"`
		LongitudeI int32 `json:"longitude_i"`
		Time       int64 `json:"time"`
	} `json:"payload"`
}

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

// metricsFromProto converts primitive fields of a protobuf message into
// Telemetry entries. Field names become the DataType.
func metricsFromProto(out *[]Telemetry, nodeID string, msg proto.Message, ts time.Time) {
	m := msg.ProtoReflect()
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !m.Has(fd) || fd.IsList() || fd.IsMap() || fd.Kind() == protoreflect.MessageKind {
			return true
		}
		t := Telemetry{NodeID: nodeID, DataType: fd.JSONName(), Timestamp: ts}
		switch fd.Kind() {
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			t.Value = v.Float()
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
			protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			t.Value = float64(v.Int())
		case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
			t.Value = float64(v.Uint())
		case protoreflect.BoolKind:
			if v.Bool() {
				t.Value = 1
			} else {
				t.Value = 0
			}
		default:
			return true
		}
		*out = append(*out, t)
		return true
	})
}

// telemetryFromProto converts a protobuf Telemetry message into Telemetry entries.
func telemetryFromProto(nodeID string, tm *mpb.Telemetry) []Telemetry {
	ts := time.Now()
	if tm.GetTime() != 0 {
		ts = time.Unix(int64(tm.GetTime()), 0)
	}
	var out []Telemetry
	if dm := tm.GetDeviceMetrics(); dm != nil {
		metricsFromProto(&out, nodeID, dm, ts)
	}
	if em := tm.GetEnvironmentMetrics(); em != nil {
		metricsFromProto(&out, nodeID, em, ts)
	}
	if aq := tm.GetAirQualityMetrics(); aq != nil {
		metricsFromProto(&out, nodeID, aq, ts)
	}
	if pm := tm.GetPowerMetrics(); pm != nil {
		metricsFromProto(&out, nodeID, pm, ts)
	}
	if ls := tm.GetLocalStats(); ls != nil {
		metricsFromProto(&out, nodeID, ls, ts)
	}
	if hm := tm.GetHealthMetrics(); hm != nil {
		metricsFromProto(&out, nodeID, hm, ts)
	}
	if host := tm.GetHostMetrics(); host != nil {
		metricsFromProto(&out, nodeID, host, ts)
	}
	return out
}

// decodeProto attempts to decode Meshtastic protobuf payloads such as
// ServiceEnvelope, Telemetry or MapReport. It returns true if the message was
// recognized and stored.
func decodeProto(store *Store, topic string, payload []byte) bool {
	if d, ok := decodeProtoMessage(topic, payload); ok {
		for _, t := range d.Telemetry {
			store.Add(t)
		}
		if d.NodeInfo != nil {
			store.SetNodeInfo(*d.NodeInfo)
		}
		return true
	}
	return false
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
		dec, err := DecodeMessage(m.Topic(), string(m.Payload()))
		if err != nil {
			log.Printf("mqtt decode: %v", err)
			return
		}
		for _, t := range dec.Telemetry {
			log.Printf("mqtt: message from %s type=%s value=%f", t.NodeID, t.DataType, t.Value)
			store.Add(t)
		}
		if dec.NodeInfo != nil {
			store.SetNodeInfo(*dec.NodeInfo)
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
