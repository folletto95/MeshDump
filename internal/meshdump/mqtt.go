package meshdump

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mpb "github.com/meshtastic/go/generated"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	var env mpb.ServiceEnvelope
	if err := proto.Unmarshal(payload, &env); err == nil {
		if pkt := env.GetPacket(); pkt != nil {
			id := fmt.Sprintf("%08x", pkt.GetFrom())
			if data := pkt.GetDecoded(); data != nil {
				switch data.GetPortnum() {
				case mpb.PortNum_TELEMETRY_APP:
					var tm mpb.Telemetry
					if err := proto.Unmarshal(data.GetPayload(), &tm); err == nil {
						for _, t := range telemetryFromProto(id, &tm) {
							store.Add(t)
						}
						return true
					}
				case mpb.PortNum_NODEINFO_APP:
					var ni mpb.NodeInfo
					if err := proto.Unmarshal(data.GetPayload(), &ni); err == nil {
						info := NodeInfo{ID: fmt.Sprintf("%08x", ni.GetNum())}
						if u := ni.GetUser(); u != nil {
							info.LongName = u.GetLongName()
							info.ShortName = u.GetShortName()
						}
						store.SetNodeInfo(info)
						return true
					}
				}
			}
		}
	}

	// MapReport is published directly without ServiceEnvelope.
	var mr pproto.MapReport
	if err := proto.Unmarshal(payload, &mr); err == nil {
		if id, ok := nodeIDFromTopic(topic); ok {
			info := NodeInfo{ID: id, LongName: mr.GetLongName(), ShortName: mr.GetShortName(), Firmware: mr.GetFirmwareVersion()}
			store.SetNodeInfo(info)
			return true
		}
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

		if decodeProto(store, m.Topic(), m.Payload()) {
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
