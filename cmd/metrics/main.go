package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

const (
	dailyPath    = "/var/lib/xray-usage/daily.json"
	snapshotPath = "/var/lib/xray-usage/snapshots.json"
)

type Usage struct {
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

type DailyData map[string]map[string]Usage

type Snapshot struct {
	Upload    uint64 `json:"upload"`
	Download  uint64 `json:"download"`
	UpdatedAt int64  `json:"updated_at"`
}

type TimeSeriesPoint struct {
	Time     string `json:"time"`
	User     string `json:"user"`
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

func loadDaily() (DailyData, error) {
	var data DailyData

	b, err := os.ReadFile(dailyPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func loadSnapshots() (map[string]Snapshot, error) {
	var data map[string]Snapshot

	b, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func handlerSnapshots(w http.ResponseWriter, r *http.Request) {
	data, err := loadSnapshots()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handlerUsers(w http.ResponseWriter, r *http.Request) {

	data, err := loadDaily()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type UserUsage struct {
		User     string `json:"user"`
		Upload   uint64 `json:"upload"`
		Download uint64 `json:"download"`
	}

	totals := make(map[string]*UserUsage)

	for _, users := range data {
		for email, usage := range users {

			if _, ok := totals[email]; !ok {
				totals[email] = &UserUsage{
					User: email,
				}
			}

			totals[email].Upload += usage.Upload
			totals[email].Download += usage.Download
		}
	}

	var result []UserUsage

	for _, u := range totals {
		result = append(result, *u)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].User < result[j].User
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handlerTimeseries(w http.ResponseWriter, r *http.Request) {

	data, err := loadDaily()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []TimeSeriesPoint

	var days []string
	for day := range data {
		days = append(days, day)
	}
	sort.Strings(days)

	for _, day := range days {

		t, err := time.Parse("2006-01-02", day)
		if err != nil {
			continue
		}

		users := data[day]

		for email, usage := range users {

			result = append(result, TimeSeriesPoint{
				Time:     t.Format(time.RFC3339),
				User:     email,
				Upload:   usage.Upload,
				Download: usage.Download,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {

	log.Println("Metrics server starting on :8085")

	http.HandleFunc("/usage/users", handlerUsers)
	http.HandleFunc("/usage/snapshots", handlerSnapshots)
	http.HandleFunc("/usage/timeseries", handlerTimeseries)

	log.Fatal(http.ListenAndServe(":8085", nil))
}