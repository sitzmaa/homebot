package storage

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Chore struct {
	ID          int
	Description string
	SubChores   []SubChore
	CompletedAt time.Time
	CompletedBy string
}

type SubChore struct {
	ID          int
	Description string
	CompletedAt time.Time
	CompletedBy string
}

type Task struct {
	ID          int
	Description string
}

type Reminder struct {
	ID        int
	Frequency string
	NextRun   time.Time
	Message   string
	ChannelID string
}

// Init initializes the SQLite database and tables
func Init(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	db.Exec("PRAGMA foreign_keys = ON")

	// Create tables if not exist
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT NOT NULL,
			completedAt DATETIME,
			completedBy TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS subchores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parentID INTEGER,
			description TEXT NOT NULL,
			completedAt DATETIME,
			completedBy TEXT,
			FOREIGN KEY(parentID) REFERENCES chores(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			frequency TEXT NOT NULL,
			nextRun DATETIME NOT NULL,
			message TEXT NOT NULL,
			channelID TEXT NOT NULL
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the database connection
func Close() error {
	return db.Close()
}

// AddChore inserts a new chore and returns its ID
func AddChore(desc string) (int, error) {
	res, err := db.Exec(`INSERT INTO chores(description) VALUES(?)`, desc)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// ListChores returns all chores with their subchores
func ListChores() ([]Chore, error) {
	rows, err := db.Query("SELECT id, description, completedAt, completedBy FROM chores")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Chore
	for rows.Next() {
		var c Chore
		var ct sql.NullTime
		rows.Scan(&c.ID, &c.Description, &ct, &c.CompletedBy)
		if ct.Valid {
			c.CompletedAt = ct.Time
		}
		// Load subchores
		subRows, _ := db.Query("SELECT id, description, completedAt, completedBy FROM subchores WHERE parentID=?", c.ID)
		for subRows.Next() {
			var s SubChore
			var st sql.NullTime
			subRows.Scan(&s.ID, &s.Description, &st, &s.CompletedBy)
			if st.Valid {
				s.CompletedAt = st.Time
			}
			c.SubChores = append(c.SubChores, s)
		}
		subRows.Close()
		out = append(out, c)
	}
	return out, nil
}

// AddSubChore adds a subchore under a parent chore
func AddSubChore(parentID, desc string) (int, error) {
	res, err := db.Exec(`INSERT INTO subchores(parentID, description) VALUES(?, ?)`, parentID, desc)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// CompleteChore marks a chore or subchore as done
func CompleteChore(target, user string) error {
	now := time.Now()
	if strings.Contains(target, ".") {
		parts := strings.SplitN(target, ".", 2)
		_, err := db.Exec(`UPDATE subchores SET completedAt=?, completedBy=? WHERE parentID=? AND id=?`, now, user, parts[0], parts[1])
		return err
	}
	_, err := db.Exec(`UPDATE chores SET completedAt=?, completedBy=? WHERE id=?`, now, user, target)
	return err
}

// PruneChores deletes chores completed more than 72h ago
func PruneChores() error {
	cutoff := time.Now().Add(-72 * time.Hour)
	_, err := db.Exec(`DELETE FROM chores WHERE completedAt IS NOT NULL AND completedAt < ?`, cutoff)
	return err
}

// AddTask creates a new task
func AddTask(desc string) (int, error) {
	res, err := db.Exec(`INSERT INTO tasks(description) VALUES(?)`, desc)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// RemoveTask deletes a task by ID
func RemoveTask(id int) error {
	_, err := db.Exec(`DELETE FROM tasks WHERE id=?`, id)
	return err
}

// ListTasks returns all pending tasks
func ListTasks() ([]Task, error) {
	rows, err := db.Query("SELECT id, description FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Description)
		out = append(out, t)
	}
	return out, nil
}

// AddReminder schedules a new reminder
func AddReminder(freq, message, channelID string) (Reminder, error) {
	now := time.Now()
	var next time.Time
	switch freq {
	case "daily":
		next = now.Add(24 * time.Hour)
	case "weekly":
		next = now.Add(7 * 24 * time.Hour)
	case "monthly":
		next = now.AddDate(0, 1, 0)
	default:
		return Reminder{}, errors.New("unsupported frequency, use daily, weekly, or monthly")
	}
	res, err := db.Exec(`INSERT INTO reminders(frequency, nextRun, message, channelID) VALUES(?,?,?,?)`, freq, next, message, channelID)
	if err != nil {
		return Reminder{}, err
	}
	id, _ := res.LastInsertId()
	return Reminder{ID: int(id), Frequency: freq, NextRun: next, Message: message, ChannelID: channelID}, nil
}

// GetDueReminders returns reminders whose nextRun ≤ now
func GetDueReminders(now time.Time) ([]Reminder, error) {
	rows, err := db.Query(`SELECT id, frequency, nextRun, message, channelID FROM reminders WHERE nextRun <= ?`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Reminder
	for rows.Next() {
		var r Reminder
		rows.Scan(&r.ID, &r.Frequency, &r.NextRun, &r.Message, &r.ChannelID)
		out = append(out, r)
	}
	return out, nil
}

// UpdateReminderNext advances or deletes a reminder after firing
func UpdateReminderNext(id int) error {
	var freq string
	var next time.Time
	err := db.QueryRow(`SELECT frequency, nextRun FROM reminders WHERE id = ?`, id).Scan(&freq, &next)
	if err != nil {
		return err
	}
	var newNext time.Time
	switch freq {
	case "daily":
		newNext = next.Add(24 * time.Hour)
	case "weekly":
		newNext = next.Add(7 * 24 * time.Hour)
	case "monthly":
		newNext = next.AddDate(0, 1, 0)
	default:
		// one-shot: remove
		_, err := db.Exec(`DELETE FROM reminders WHERE id = ?`, id)
		return err
	}
	_, err = db.Exec(`UPDATE reminders SET nextRun = ? WHERE id = ?`, newNext, id)
	return err
}
