package main

import (
	"log"
	"os"

	"github.com/lbenicio/pocketid-sidecar-listmonk/internal/config"
	"github.com/lbenicio/pocketid-sidecar-listmonk/internal/syncer"
)

func main() {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	s := syncer.New(cfg)
	stats, err := s.Run()
	if err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	log.Printf("--- sync complete ---")
	log.Printf("created: %d", stats.Created)
	log.Printf("updated: %d", stats.Updated)
	log.Printf("deleted: %d", stats.Deleted)
	if stats.Errors > 0 {
		log.Printf("errors:  %d", stats.Errors)
		os.Exit(1)
	}
}
