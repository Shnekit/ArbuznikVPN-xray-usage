package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const (
	dailyPath     = "/var/lib/xray-usage/daily.json"
	snapshotPath  = "/var/lib/xray-usage/snapshots.json"
)

//
// -------------------- DATA MODELS --------------------
//

type DailyRecord struct {
	Day      int64  `json:"day"`
	Email    string `json:"email"`
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

type Snapshot struct {
	UploadBytes   uint64 `json:"upload_bytes"`
	DownloadBytes uint64 `json:"download_bytes"`
}

type DailyData []DailyRecord

//
// -------------------- LOADERS --------------------
//

func loadDaily(path string) (DailyData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data DailyData
	err = json.Unmarshal(b, &data)
	return data, err
}

func loadSnapshots(path string) (map[string]Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]Snapshot
	err = json.Unmarshal(b, &data)
	return data, err
}

//
// -------------------- HANDLERS --------------------
//

// Grafana-friendly endpoint:
// returns ALL users in one request
func handlerPerUserUsage(w http.ResponseWriter, r *http.Request) {
	data, err := loadDaily(dailyPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type agg struct {
		Upload   uint64
		Download uint64
	}

	result := make(map[string]agg)

	// aggregate per user
	for _, d := range data {
		a := result[d.Email]
		a.Upload += d.Upload
		a.Download += d.Download
		result[d.Email] = a
	}

	// convert map -> array (Grafana-friendly)
	out := make([]map[string]interface{}, 0, len(result))

	for user, v := range result {
		out = append(out, map[string]interface{}{
			"user":     user,
			"upload":   v.Upload,
			"download": v.Download,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// Optional: raw snapshots (useful for debugging or admin panel)
func handlerSnapshots(w http.ResponseWriter, r *http.Request) {
	data, err := loadSnapshots(snapshotPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handlerTimeseries(w http.ResponseWriter, r *http.Request) {
	data, err := loadDaily(dailyPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out := make([]map[string]interface{}, 0, len(data))

	for _, d := range data {
		out = append(out, map[string]interface{}{
			"time":   time.Unix(d.Day, 0).UTC().Format(time.RFC3339),
			"user":   d.Email,
			"upload": d.Upload,
			"download": d.Download,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

//
// -------------------- MAIN --------------------
//

func main() {
	log.Println("Metrics server starting on :8085")

	http.HandleFunc("/usage/users", handlerPerUserUsage)
	http.HandleFunc("/usage/snapshots", handlerSnapshots)
	http.HandleFunc("/usage/timeseries", handlerTimeseries)

	log.Fatal(http.ListenAndServe(":8085", nil))
}