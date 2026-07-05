package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type DailyData map[string]map[string]struct {
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

type Snapshot struct {
	UploadBytes   uint64 `json:"upload_bytes"`
	DownloadBytes uint64 `json:"download_bytes"`
}

func loadDaily(path string) (DailyData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data DailyData
	err = json.Unmarshal(b, &data)
	return data, err
}

func handlerDaily(w http.ResponseWriter, r *http.Request) {
	data, err := loadDaily("/var/lib/xray-usage/daily.json")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// flatten per user totals (last day simple version)
	result := map[string]uint64{}

	for _, user := range data {
		for _, d := range user {
			result["upload"] += d.Upload
			result["download"] += d.Download
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handlerSnapshots(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile("/var/lib/xray-usage/snapshots.json")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var data map[string]Snapshot
	json.Unmarshal(b, &data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	log.Println("Metrics server starting on :8085")

	http.HandleFunc("/usage/daily", handlerDaily)
	http.HandleFunc("/usage/snapshots", handlerSnapshots)

	log.Fatal(http.ListenAndServe(":8085", nil))
}