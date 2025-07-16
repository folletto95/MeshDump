package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"meshdump/internal/meshdump"
)

func loadEnv() {
	var paths []string
	// first try the working directory
	paths = append(paths, ".env")
	// also try alongside the executable
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), ".env"))
	}

	var data []byte
	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			data = d
			break
		}
	}
	if len(data) == 0 {
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
		if err := os.Setenv(key, val); err != nil {
			log.Printf("loadEnv: %v", err)
		}
	}
}

func main() {
	loadEnv()

	log.Printf("MeshDump version %s", meshdump.Version)

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

	log.Println("Starting MeshDump on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.Router()))
}
