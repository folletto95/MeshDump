package meshdump

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mpb "github.com/meshtastic/go/generated"
	"google.golang.org/protobuf/proto"
	pproto "meshdump/internal/proto"
)

// Decoded holds telemetry entries or node info extracted from a payload.
type Decoded struct {
	Telemetry []Telemetry
	NodeInfo  *NodeInfo
}

// DecodeMessage attempts to decode an MQTT payload that may contain JSON or
// protobuf encoded data. The topic is used to infer the node ID when missing.
func DecodeMessage(topic, payload string) (*Decoded, error) {
	trimmed := strings.TrimSpace(payload)
	if strings.HasPrefix(trimmed, "{") {
		if d, ok := decodeJSON(topic, []byte(trimmed)); ok {
			return d, nil
		}
	}

	raw := []byte(trimmed)
	if b, err := base64.StdEncoding.DecodeString(trimmed); err == nil {
		raw = b
	}
	if d, ok := decodeProtoMessage(topic, raw); ok {
		return d, nil
	}
	return nil, fmt.Errorf("unknown payload format")
}

func decodeJSON(topic string, data []byte) (*Decoded, bool) {
	var tel Telemetry
	if err := json.Unmarshal(data, &tel); err == nil {
		if tel.NodeID == "" {
			if id, ok := nodeIDFromTopic(topic); ok {
				tel.NodeID = id
			}
		}
		if tel.NodeID != "" {
			return &Decoded{Telemetry: []Telemetry{tel}}, true
		}
	}

	var pos jsonPosition
	if err := json.Unmarshal(data, &pos); err == nil && pos.Type == "position" {
		id := strings.TrimPrefix(pos.Sender, "!")
		if id == "" && pos.From != 0 {
			id = fmt.Sprintf("%08x", pos.From)
		}
		if id == "" {
			if tID, ok := nodeIDFromTopic(topic); ok {
				id = tID
			}
		}
		if id == "" {
			return nil, false
		}
		ts := time.Now()
		if pos.Payload.Time != 0 {
			ts = time.Unix(pos.Payload.Time, 0)
		} else if pos.Timestamp != 0 {
			ts = time.Unix(pos.Timestamp, 0)
		}
		lat := float64(pos.Payload.LatitudeI) / 1e7
		lon := float64(pos.Payload.LongitudeI) / 1e7
		return &Decoded{Telemetry: []Telemetry{
			{NodeID: id, DataType: "latitude", Value: lat, Timestamp: ts},
			{NodeID: id, DataType: "longitude", Value: lon, Timestamp: ts},
		}}, true
	}

	return nil, false
}

func decodeProtoMessage(topic string, payload []byte) (*Decoded, bool) {
	var env mpb.ServiceEnvelope
	if err := proto.Unmarshal(payload, &env); err == nil {
		if pkt := env.GetPacket(); pkt != nil {
			id := fmt.Sprintf("%08x", pkt.GetFrom())
			if data := pkt.GetDecoded(); data != nil {
				switch data.GetPortnum() {
				case mpb.PortNum_TELEMETRY_APP:
					var tm mpb.Telemetry
					if err := proto.Unmarshal(data.GetPayload(), &tm); err == nil {
						return &Decoded{Telemetry: telemetryFromProto(id, &tm)}, true
					}
				case mpb.PortNum_NODEINFO_APP:
					var ni mpb.NodeInfo
					if err := proto.Unmarshal(data.GetPayload(), &ni); err == nil {
						info := NodeInfo{ID: fmt.Sprintf("%08x", ni.GetNum())}
						if u := ni.GetUser(); u != nil {
							info.LongName = u.GetLongName()
							info.ShortName = u.GetShortName()
						}
						return &Decoded{NodeInfo: &info}, true
					}
				case mpb.PortNum_POSITION_APP:
					var pos mpb.Position
					if err := proto.Unmarshal(data.GetPayload(), &pos); err == nil {
						ts := time.Now()
						if pos.GetTime() != 0 {
							ts = time.Unix(int64(pos.GetTime()), 0)
						} else if pos.GetTimestamp() != 0 {
							ts = time.Unix(int64(pos.GetTimestamp()), 0)
						}
						lat := float64(pos.GetLatitudeI()) / 1e7
						lon := float64(pos.GetLongitudeI()) / 1e7
						tel := []Telemetry{
							{NodeID: id, DataType: "latitude", Value: lat, Timestamp: ts},
							{NodeID: id, DataType: "longitude", Value: lon, Timestamp: ts},
						}
						if alt := pos.GetAltitude(); alt != 0 {
							tel = append(tel, Telemetry{NodeID: id, DataType: "altitude", Value: float64(alt), Timestamp: ts})
						}
						return &Decoded{Telemetry: tel}, true
					}
				}
			}
		}
	}

	var mr pproto.MapReport
	if err := proto.Unmarshal(payload, &mr); err == nil {
		if id, ok := nodeIDFromTopic(topic); ok {
			info := NodeInfo{ID: id, LongName: mr.GetLongName(), ShortName: mr.GetShortName(), Firmware: mr.GetFirmwareVersion()}
			return &Decoded{NodeInfo: &info}, true
		}
	}

	return nil, false
}
