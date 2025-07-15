package meshdump

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	mpb "github.com/meshtastic/go/generated"
	"google.golang.org/protobuf/proto"
	pproto "meshdump/internal/proto"
)

func TestDecodeMessageJSON(t *testing.T) {
	tel := Telemetry{NodeID: "node1", DataType: "temperature", Value: 12.5}
	data, _ := json.Marshal(tel)
	dec, err := DecodeMessage("msh/node1", string(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(dec.Telemetry) != 1 || dec.Telemetry[0].NodeID != tel.NodeID {
		t.Fatalf("unexpected telemetry: %+v", dec.Telemetry)
	}
}

func TestDecodeMessageProto(t *testing.T) {
	batt := uint32(50)
	tm := &mpb.Telemetry{Time: 1000,
		Variant: &mpb.Telemetry_DeviceMetrics{DeviceMetrics: &mpb.DeviceMetrics{BatteryLevel: &batt}},
	}
	tmData, _ := proto.Marshal(tm)
	pkt := &mpb.MeshPacket{From: 1, PayloadVariant: &mpb.MeshPacket_Decoded{Decoded: &mpb.Data{Portnum: mpb.PortNum_TELEMETRY_APP, Payload: tmData}}}
	env := &mpb.ServiceEnvelope{Packet: pkt}
	raw, _ := proto.Marshal(env)
	enc := base64.StdEncoding.EncodeToString(raw)
	dec, err := DecodeMessage("msh/00000001", enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(dec.Telemetry) == 0 {
		t.Fatalf("no telemetry decoded")
	}
	if dec.Telemetry[0].NodeID != "00000001" {
		t.Errorf("unexpected node id: %s", dec.Telemetry[0].NodeID)
	}
}

func TestDecodeMessageMapReport(t *testing.T) {
	mr := &pproto.MapReport{LongName: "Node", FirmwareVersion: "1.0"}
	raw, _ := proto.Marshal(mr)
	enc := base64.StdEncoding.EncodeToString(raw)
	dec, err := DecodeMessage("msh/nodeX", enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dec.NodeInfo == nil || dec.NodeInfo.LongName != "Node" {
		t.Fatalf("unexpected node info: %+v", dec.NodeInfo)
	}
}
