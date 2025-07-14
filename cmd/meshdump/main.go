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

func loadEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") && len(val) >= 2 {
			val = strings.Trim(val, "\"")
		}
		os.Setenv(key, val)
	}
}

func main() {
	loadEnv(".env")
	nodesEnv := os.Getenv("NODES")
	var nodes []string
	if nodesEnv != "" {
		nodes = strings.Split(nodesEnv, ",")
	}
	log.Printf("config: nodes=%v", nodes)

	dataFile := os.Getenv("DATA_FILE")
	log.Printf("config: data file=%s", dataFile)
	store := meshdump.NewStore(dataFile)
	server := meshdump.NewServer(store)

	mqttBroker := os.Getenv("MQTT_BROKER")
	mqttTopic := os.Getenv("MQTT_TOPIC")
	if mqttTopic == "" {
		mqttTopic = "#"
	}
	mqttUser := os.Getenv("MQTT_USERNAME")
	mqttPass := os.Getenv("MQTT_PASSWORD")
	log.Printf("config: mqtt broker=%s topic=%s", mqttBroker, mqttTopic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if mqttBroker != "" {
		if err := meshdump.StartMQTT(ctx, mqttBroker, mqttTopic, mqttUser, mqttPass, store); err != nil {
			log.Fatalf("mqtt: %v", err)
		}
	}
	if len(nodes) > 0 {
		go meshdump.PollNodes(ctx, time.Minute, store, nodes)
	}

	log.Println("Starting MeshDump on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.Router()))
}
