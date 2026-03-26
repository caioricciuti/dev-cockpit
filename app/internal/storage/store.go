package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// MetricSnapshot represents a single point-in-time measurement
type MetricSnapshot struct {
	Timestamp  time.Time
	CPUAvg     float64
	MemoryPct  float64
	DiskPct    float64
	NetInRate  float64 // bytes/sec
	NetOutRate float64 // bytes/sec
}

// Store is the persistent SQLite metrics store
type Store struct {
	db *sql.DB
	mu sync.Mutex
}

// Open creates or opens the SQLite database at the given path
func Open(dbPath string) (*Store, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better read/write concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS metrics (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  INTEGER NOT NULL,
			cpu_avg    REAL NOT NULL,
			memory_pct REAL NOT NULL,
			disk_pct   REAL NOT NULL,
			net_in     REAL NOT NULL,
			net_out    REAL NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics(timestamp);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Insert writes a single metric snapshot
func (s *Store) Insert(snap MetricSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		"INSERT INTO metrics (timestamp, cpu_avg, memory_pct, disk_pct, net_in, net_out) VALUES (?, ?, ?, ?, ?, ?)",
		snap.Timestamp.Unix(),
		snap.CPUAvg,
		snap.MemoryPct,
		snap.DiskPct,
		snap.NetInRate,
		snap.NetOutRate,
	)
	return err
}

// Query retrieves metrics for the given duration, automatically downsampled
// for larger time ranges to keep result sets reasonable.
func (s *Store) Query(since time.Duration) ([]MetricSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-since).Unix()

	// Determine grouping interval for downsampling
	var query string
	switch {
	case since <= time.Hour:
		// Raw data points
		query = `SELECT timestamp, cpu_avg, memory_pct, disk_pct, net_in, net_out
			FROM metrics WHERE timestamp >= ? ORDER BY timestamp`
	case since <= 6*time.Hour:
		// 1-minute averages
		query = `SELECT (timestamp / 60) * 60, AVG(cpu_avg), AVG(memory_pct), AVG(disk_pct), AVG(net_in), AVG(net_out)
			FROM metrics WHERE timestamp >= ? GROUP BY timestamp / 60 ORDER BY timestamp / 60`
	case since <= 24*time.Hour:
		// 5-minute averages
		query = `SELECT (timestamp / 300) * 300, AVG(cpu_avg), AVG(memory_pct), AVG(disk_pct), AVG(net_in), AVG(net_out)
			FROM metrics WHERE timestamp >= ? GROUP BY timestamp / 300 ORDER BY timestamp / 300`
	default:
		// 15-minute averages
		query = `SELECT (timestamp / 900) * 900, AVG(cpu_avg), AVG(memory_pct), AVG(disk_pct), AVG(net_in), AVG(net_out)
			FROM metrics WHERE timestamp >= ? GROUP BY timestamp / 900 ORDER BY timestamp / 900`
	}

	rows, err := s.db.Query(query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var snapshots []MetricSnapshot
	for rows.Next() {
		var ts int64
		var snap MetricSnapshot
		if err := rows.Scan(&ts, &snap.CPUAvg, &snap.MemoryPct, &snap.DiskPct, &snap.NetInRate, &snap.NetOutRate); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		snap.Timestamp = time.Unix(ts, 0)
		snapshots = append(snapshots, snap)
	}

	return snapshots, rows.Err()
}

// Prune deletes data older than maxDays
func (s *Store) Prune(maxDays int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -maxDays).Unix()
	_, err := s.db.Exec("DELETE FROM metrics WHERE timestamp < ?", cutoff)
	return err
}
