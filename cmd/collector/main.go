package main

import (
	"log"

	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/collector"
	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/storage"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("========== XRAY USAGE COLLECTOR STARTING ==========")

	log.Println("[1/4] Opening file...")

	file, err := storage.Open("/var/lib/xray-usage")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	log.Println("[2/4] File opened successfully.")

	log.Println("[3/4] Creating collector...")
	c := collector.New(file)
	log.Println("Collector created.")

	log.Println("[4/4] Running collector...")

	if err := c.Run(); err != nil {
		log.Fatalf("Collector failed: %v", err)
	}

	log.Println("Collector finished successfully.")
	log.Println("========== DONE ==========")
}