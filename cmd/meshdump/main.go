package main

import (
	"log"
	"net/http"

	"meshdump/internal/meshdump"
)

func main() {
	store := meshdump.NewStore()
	server := meshdump.NewServer(store)

	log.Println("Starting MeshDump on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.Router()))
}
