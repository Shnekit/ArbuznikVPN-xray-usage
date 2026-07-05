package collector

import (
	"log"
	"time"

	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/storage"
	"github.com/Shnekit/ArbuznikVPN-xray-usage/internal/xray"
)

type Collector struct {
	store *storage.Store
}

func New(store *storage.Store) *Collector {
	return &Collector{
		store: store,
	}
}

func (c *Collector) Run() error {

	now := time.Now().UTC()
	day := now.Format("2006-01-02")

	log.Println("[collector] reading xray stats...")

	currentStats, err := xray.ReadStats()
	if err != nil {
		return err
	}

	for _, current := range currentStats {

		// ----------------------------
		// GET OLD SNAPSHOT
		// ----------------------------
		snap, exists := c.store.GetSnapshot(current.Email)

		if !exists {
			// First time seen user → initialize baseline
			c.store.SetSnapshot(current.Email, storage.Snapshot{
				Upload:    current.UploadBytes,
				Download:  current.DownloadBytes,
				UpdatedAt: now.Unix(),
			})
			continue
		}

		// ----------------------------
		// COMPUTE DELTA
		// ----------------------------
		uploadDelta := computeDelta(current.UploadBytes, snap.Upload)
		downloadDelta := computeDelta(current.DownloadBytes, snap.Download)

		// ----------------------------
		// UPDATE DAILY USAGE
		// ----------------------------
		c.store.AddDaily(day, current.Email, uploadDelta, downloadDelta)

		// ----------------------------
		// UPDATE SNAPSHOT
		// ----------------------------
		c.store.SetSnapshot(current.Email, storage.Snapshot{
			Upload:    current.UploadBytes,
			Download:  current.DownloadBytes,
			UpdatedAt: now.Unix(),
		})
	}

	// ----------------------------
	// CLEAN OLD DATA
	// ----------------------------
	c.store.Cleanup()

	// ----------------------------
	// PERSIST TO DISK (IMPORTANT)
	// ----------------------------
	if err := c.store.Save(); err != nil {
		return err
	}

	log.Println("[collector] finished successfully")

	return nil
}

func computeDelta(current, previous uint64) uint64 {

	if current >= previous {
		return current - previous
	}

	// counter reset (xray restart)
	return current
}