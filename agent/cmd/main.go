package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark-chris/devtools-sync/agent/internal/config"
)

const version = "0.1.0"

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("devtools-sync-agent version %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("devtools-sync-agent starting (version %s)", version)
	log.Printf("Server URL: %s", cfg.ServerURL)
	log.Println("Agent initialized successfully")
}
