package main

import (
	"log"

	"github.com/yourusername/xray-usage-collector/internal/collector"
	"github.com/yourusername/xray-usage-collector/internal/database"
)

func main() {

	// Open SQLite DB (file will be created if not exists)
	db, err := database.Open("/var/lib/xray-usage/usage.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	c := collector.New(db)

	if err := c.Run(); err != nil {
		log.Fatalf("collector failed: %v", err)
	}

	log.Println("collector run completed successfully")
}
