package model

import "time"

// UserStats represents the current counters returned by Xray.
type UserStats struct {
	Email         string
	UploadBytes   uint64
	DownloadBytes uint64
}

// Snapshot stores the latest counters we have persisted.
type Snapshot struct {
	Email         string
	UploadBytes   uint64
	DownloadBytes uint64
	UpdatedAt     time.Time
}

// DailyUsage stores accumulated traffic for one day.
type DailyUsage struct {
	Date          string
	Email         string
	UploadBytes   uint64
	DownloadBytes uint64
}
