package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Email struct {
	MessageID      string `json:"message_id"`
	From           string `json:"from"`
	Title          string `json:"title"`
	Subject        string `json:"subject"`
	Link           string `json:"link"`
	Folder         string `json:"folder"`
	Timestamp      string `json:"timestamp"`
	SnippetPreview string `json:"snippet_preview"`
	IsRead         bool   `json:"is_read"`
	CreatedAt      string `json:"created_at"`
}

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	sqlDb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := sqlDb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{db: sqlDb}

	// Create table
	if err := db.CreateTable(); err != nil {
		return nil, err
	}

	return db, nil
}

func (d *Database) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS emails (
		message_id TEXT PRIMARY KEY,
		from_addr TEXT,
		title TEXT,
		subject TEXT,
		link TEXT,
		folder TEXT,
		timestamp TEXT,
		snippet_preview TEXT,
		is_read BOOLEAN,
		created_at TEXT
	)
	`

	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_folder ON emails(folder)`,
		`CREATE INDEX IF NOT EXISTS idx_timestamp ON emails(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_title ON emails(title)`,
	}

	for _, idx := range indexes {
		if _, err := d.db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func (d *Database) IndexEmail(email *Email) error {
	query := `
	INSERT OR REPLACE INTO emails (message_id, from_addr, title, subject, link, folder, timestamp, snippet_preview, is_read, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, email.MessageID, email.From, email.Title, email.Subject, email.Link, email.Folder, email.Timestamp, email.SnippetPreview, email.IsRead, email.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to index email: %w", err)
	}

	return nil
}

func (d *Database) IndexEmails(emails []*Email) error {
	for _, email := range emails {
		if err := d.IndexEmail(email); err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) SearchEmails(query string, page int, pageSize int) ([]Email, int, error) {
	offset := (page - 1) * pageSize

	// Count total
	countQuery := `SELECT COUNT(*) FROM emails WHERE title LIKE ? OR subject LIKE ? OR from_addr LIKE ?`
	searchTerm := "%" + query + "%"
	var total int
	err := d.db.QueryRow(countQuery, searchTerm, searchTerm, searchTerm).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count emails: %w", err)
	}

	// Get paginated results
	selectQuery := `
	SELECT message_id, from_addr, title, subject, link, folder, timestamp, snippet_preview, is_read, created_at
	FROM emails
	WHERE title LIKE ? OR subject LIKE ? OR from_addr LIKE ?
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := d.db.Query(selectQuery, searchTerm, searchTerm, searchTerm, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search emails: %w", err)
	}
	defer rows.Close()

	var emails []Email
	for rows.Next() {
		var email Email
		err := rows.Scan(&email.MessageID, &email.From, &email.Title, &email.Subject, &email.Link, &email.Folder, &email.Timestamp, &email.SnippetPreview, &email.IsRead, &email.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, total, nil
}

func (d *Database) DeleteEmail(messageID string) error {
	query := `DELETE FROM emails WHERE message_id = ?`
	_, err := d.db.Exec(query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}
	return nil
}

func (d *Database) GetStats() (map[string]interface{}, error) {
	var totalEmails int
	query := `SELECT COUNT(*) FROM emails`
	err := d.db.QueryRow(query).Scan(&totalEmails)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_emails": totalEmails,
	}

	return stats, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
