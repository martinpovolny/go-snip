package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store is a SQLite-backed ring buffer for monitor check results.
// Each monitor keeps at most `keep` rows; older rows are pruned on every insert.
const defaultKeep = 2016 // ~1 week at 5-min intervals

type Store struct {
	db   *sql.DB
	keep int
}

func NewStore(path string, keep int) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS monitor_checks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			monitor     TEXT    NOT NULL,
			checked_at  INTEGER NOT NULL,
			up          INTEGER NOT NULL,
			status_code INTEGER,
			response_ms INTEGER,
			error       TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_mc ON monitor_checks (monitor, checked_at DESC);

		CREATE TABLE IF NOT EXISTS notifications (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			sent_at  INTEGER NOT NULL,
			monitor  TEXT    NOT NULL,
			event    TEXT    NOT NULL,
			title    TEXT    NOT NULL,
			message  TEXT    NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_notif ON notifications (sent_at DESC);

		CREATE TABLE IF NOT EXISTS monitor_alert_state (
			monitor     TEXT    PRIMARY KEY,
			consecutive INTEGER NOT NULL DEFAULT 0,
			alerted     INTEGER NOT NULL DEFAULT 0
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{db: db, keep: keep}, nil
}

func (s *Store) Record(ms MonitorStatus) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO monitor_checks (monitor, checked_at, up, status_code, response_ms, error)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		ms.Name, ms.CheckedAt, boolInt(ms.Up), ms.StatusCode, ms.ResponseMs, nullStr(ms.Error),
	)
	if err != nil {
		return err
	}

	// Prune rows beyond keep limit for this monitor
	_, err = tx.Exec(
		`DELETE FROM monitor_checks
		 WHERE monitor = ? AND id NOT IN (
		   SELECT id FROM monitor_checks WHERE monitor = ? ORDER BY checked_at DESC LIMIT ?
		 )`,
		ms.Name, ms.Name, s.keep,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) History(monitor string, limit int) ([]MonitorStatus, error) {
	rows, err := s.db.Query(
		`SELECT monitor, checked_at, up, status_code, response_ms, error
		 FROM monitor_checks
		 WHERE monitor = ?
		 ORDER BY checked_at DESC
		 LIMIT ?`,
		monitor, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MonitorStatus
	for rows.Next() {
		var ms MonitorStatus
		var up int
		var code, rtt sql.NullInt64
		var errStr sql.NullString
		if err := rows.Scan(&ms.Name, &ms.CheckedAt, &up, &code, &rtt, &errStr); err != nil {
			return nil, err
		}
		ms.Up = up == 1
		if code.Valid {
			ms.StatusCode = int(code.Int64)
		}
		if rtt.Valid {
			ms.ResponseMs = rtt.Int64
		}
		if errStr.Valid {
			ms.Error = errStr.String
		}
		out = append(out, ms)
	}
	return out, rows.Err()
}

type NotificationRecord struct {
	ID      int64  `json:"id"`
	SentAt  int64  `json:"sent_at"`
	Monitor string `json:"monitor"`
	Event   string `json:"event"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (s *Store) RecordNotification(monitor, event, title, message string) error {
	_, err := s.db.Exec(
		`INSERT INTO notifications (sent_at, monitor, event, title, message) VALUES (?, ?, ?, ?, ?)`,
		time.Now().Unix(), monitor, event, title, message,
	)
	return err
}

func (s *Store) Notifications(limit int) ([]NotificationRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, sent_at, monitor, event, title, message
		 FROM notifications ORDER BY sent_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NotificationRecord
	for rows.Next() {
		var n NotificationRecord
		if err := rows.Scan(&n.ID, &n.SentAt, &n.Monitor, &n.Event, &n.Title, &n.Message); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// LoadAlertState returns persisted alert state. If no row exists it infers
// consecutive failure count from recent monitor_checks history so a restart
// mid-outage doesn't silently reset the counter.
func (s *Store) LoadAlertState(monitor string, threshold int) (consecutive int, alerted bool) {
	var a int
	err := s.db.QueryRow(
		`SELECT consecutive, alerted FROM monitor_alert_state WHERE monitor = ?`, monitor,
	).Scan(&consecutive, &a)
	if err == nil {
		return consecutive, a == 1
	}
	// No persisted row — infer from recent check history.
	rows, err := s.db.Query(
		`SELECT up FROM monitor_checks WHERE monitor = ? ORDER BY checked_at DESC LIMIT ?`,
		monitor, threshold*2,
	)
	if err != nil {
		return 0, false
	}
	defer rows.Close()
	for rows.Next() {
		var up int
		if err := rows.Scan(&up); err != nil {
			break
		}
		if up == 0 {
			consecutive++
		} else {
			break
		}
	}
	return consecutive, consecutive >= threshold
}

func (s *Store) SaveAlertState(monitor string, consecutive int, alerted bool) error {
	_, err := s.db.Exec(
		`INSERT INTO monitor_alert_state (monitor, consecutive, alerted)
		 VALUES (?, ?, ?)
		 ON CONFLICT(monitor) DO UPDATE SET consecutive=excluded.consecutive, alerted=excluded.alerted`,
		monitor, consecutive, boolInt(alerted),
	)
	return err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
