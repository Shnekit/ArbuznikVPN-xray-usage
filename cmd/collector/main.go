package main

import (
	"log"

	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/collector"
	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/database"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("========== XRAY USAGE COLLECTOR STARTING ==========")

	log.Println("[1/4] Opening SQLite database...")

	db, err := database.Open("/var/lib/xray-usage/usage.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		log.Println("Closing database...")
		db.Close()
	}()

	log.Println("[2/4] Database opened successfully.")

	log.Println("[3/4] Creating collector...")
	c := collector.New(db)
	log.Println("Collector created.")

	log.Println("[4/4] Running collector...")

	if err := c.Run(); err != nil {
		log.Fatalf("Collector failed: %v", err)
	}

	log.Println("Collector finished successfully.")
	log.Println("========== DONE ==========")
}