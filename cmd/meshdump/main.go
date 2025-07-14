package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"meshdump/internal/meshdump"
)

func main() {
	nodesEnv := os.Getenv("NODES")
	var nodes []string
	if nodesEnv != "" {
		nodes = strings.Split(nodesEnv, ",")
	}

	dataFile := os.Getenv("DATA_FILE")
	store := meshdump.NewStore(dataFile)
	server := meshdump.NewServer(store)

	mqttBroker := os.Getenv("MQTT_BROKER")
	mqttTopic := os.Getenv("MQTT_TOPIC")
	if mqttTopic == "" {
		mqttTopic = "telemetry/#"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if mqttBroker != "" {
		if err := meshdump.StartMQTT(ctx, mqttBroker, mqttTopic, store); err != nil {
			log.Fatalf("mqtt: %v", err)
		}
	}
	if len(nodes) > 0 {
		go meshdump.PollNodes(ctx, time.Minute, store, nodes)
	}

	log.Println("Starting MeshDump on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.Router()))
}
