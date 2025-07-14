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

	store := meshdump.NewStore()
	server := meshdump.NewServer(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if len(nodes) > 0 {
		go meshdump.PollNodes(ctx, time.Minute, store, nodes)
	}

	log.Println("Starting MeshDump on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.Router()))
}
