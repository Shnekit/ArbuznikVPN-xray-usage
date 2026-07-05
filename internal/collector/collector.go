package collector

import (
	"time"

	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/database"
	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/model"
	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/xray"
)

type Collector struct {
	db *database.Database
}

func New(db *database.Database) *Collector {
	return &Collector{
		db: db,
	}
}

func (c *Collector) Run() error {

	// Read current Xray counters.
	currentStats, err := xray.ReadStats()
	if err != nil {
		return err
	}

	// Load previous snapshots.
	snapshots, err := c.db.LoadSnapshots()
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	for _, current := range currentStats {

		snapshot, exists := snapshots[current.Email]

		// FIRST TIME SEEN USER → just create baseline, no usage counted
		if !exists {
    			_ = c.db.SaveSnapshot(model.Snapshot{
        			Email:         current.Email,
        			UploadBytes:   current.UploadBytes,
        			DownloadBytes: current.DownloadBytes,
        			UpdatedAt:     now,
    			})
    			continue
		}

		// Normal case: compute delta
		uploadDelta := computeDelta(
    			current.UploadBytes,
    			snapshot.UploadBytes,
		)

		downloadDelta := computeDelta(
    			current.DownloadBytes,
    			snapshot.DownloadBytes,
		)

		// Store daily usage
		err = c.db.AddDailyUsage(
    			current.Email,
    			uploadDelta,
    			downloadDelta,
		)
		if err != nil {
    			return err
		}


		// Save latest snapshot.
		err = c.db.SaveSnapshot(model.Snapshot{
			Email:         current.Email,
			UploadBytes:   current.UploadBytes,
			DownloadBytes: current.DownloadBytes,
			UpdatedAt:     now,
		})
		if err != nil {
			return err
		}
	}

	// Remove data older than 30 days.
	return c.db.DeleteOldHistory()
}

func computeDelta(current, previous uint64) uint64 {

	// Normal case.
	if current >= previous {
		return current - previous
	}

	// Xray restarted and counters reset.
	return current
}
