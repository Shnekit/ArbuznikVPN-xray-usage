package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/yourusername/xray-usage-collector/internal/model"
)

type Database struct {
	db *sql.DB
}

func Open(path string) (*Database, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, err
	}

	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		return nil, err
	}

	d := &Database{
		db: db,
	}

	if err := d.createTables(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) createTables() error {

	schema := `
CREATE TABLE IF NOT EXISTS snapshots (
	email TEXT PRIMARY KEY,
	upload_bytes INTEGER NOT NULL,
	download_bytes INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS daily_usage (
	day INTEGER NOT NULL,
	email TEXT NOT NULL,
	upload_bytes INTEGER NOT NULL,
	download_bytes INTEGER NOT NULL,

	PRIMARY KEY(day, email)
);

CREATE INDEX IF NOT EXISTS idx_daily_usage_day
	ON daily_usage(day);

CREATE INDEX IF NOT EXISTS idx_daily_usage_email
	ON daily_usage(email);
`

	_, err := d.db.Exec(schema)
	return err
}

// Returns the snapshot for a user.
// If it doesn't exist, returns (nil, nil).
func (d *Database) GetSnapshot(email string) (*model.Snapshot, error) {

	row := d.db.QueryRow(`
SELECT
	email,
	upload_bytes,
	download_bytes,
	updated_at
FROM snapshots
WHERE email = ?`,
		email,
	)

	var snapshot model.Snapshot
	var updated int64

	err := row.Scan(
		&snapshot.Email,
		&snapshot.UploadBytes,
		&snapshot.DownloadBytes,
		&updated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	snapshot.UpdatedAt = time.Unix(updated, 0)

	return &snapshot, nil
}

func (d *Database) SaveSnapshot(snapshot model.Snapshot) error {

	_, err := d.db.Exec(`
INSERT INTO snapshots (
	email,
	upload_bytes,
	download_bytes,
	updated_at
)
VALUES (?, ?, ?, ?)

ON CONFLICT(email)
DO UPDATE SET
	upload_bytes = excluded.upload_bytes,
	download_bytes = excluded.download_bytes,
	updated_at = excluded.updated_at;
`,
		snapshot.Email,
		snapshot.UploadBytes,
		snapshot.DownloadBytes,
		snapshot.UpdatedAt.Unix(),
	)

	return err
}

// Adds today's delta to the accumulated daily usage.
func (d *Database) AddDailyUsage(
	email string,
	uploadDelta uint64,
	downloadDelta uint64,
) error {

	day := beginningOfToday()

	_, err := d.db.Exec(`
INSERT INTO daily_usage (
	day,
	email,
	upload_bytes,
	download_bytes
)
VALUES (?, ?, ?, ?)

ON CONFLICT(day, email)
DO UPDATE SET

	upload_bytes =
		daily_usage.upload_bytes + excluded.upload_bytes,

	download_bytes =
		daily_usage.download_bytes + excluded.download_bytes;
`,
		day,
		email,
		uploadDelta,
		downloadDelta,
	)

	return err
}

// Delete history older than 30 days.
func (d *Database) DeleteOldHistory() error {

	cutoff := beginningOfToday() - 30*24*60*60

	_, err := d.db.Exec(`
DELETE
FROM daily_usage
WHERE day < ?`,
		cutoff,
	)

	return err
}

func beginningOfToday() int64 {

	now := time.Now().UTC()

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	return today.Unix()
}

func (d *Database) LoadSnapshots() (map[string]model.Snapshot, error) {

	rows, err := d.db.Query(`
SELECT
	email,
	upload_bytes,
	download_bytes,
	updated_at
FROM snapshots
`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	snapshots := make(map[string]model.Snapshot)

	for rows.Next() {

		var snapshot model.Snapshot
		var updated int64

		err := rows.Scan(
			&snapshot.Email,
			&snapshot.UploadBytes,
			&snapshot.DownloadBytes,
			&updated,
		)
		if err != nil {
			return nil, err
		}

		snapshot.UpdatedAt = time.Unix(updated, 0)

		snapshots[snapshot.Email] = snapshot
	}

	return snapshots, nil
}

func (d *Database) DebugPrint() error {

	rows, err := d.db.Query(`
SELECT
	email,
	upload_bytes,
	download_bytes
FROM snapshots
ORDER BY email
`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {

		var email string
		var up uint64
		var down uint64

		rows.Scan(&email, &up, &down)

		fmt.Printf(
			"%s up=%d down=%d\n",
			email,
			up,
			down,
		)
	}

	return nil
}
