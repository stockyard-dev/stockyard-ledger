package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

// Transaction is a single ledger entry. Amount is stored as integer cents
// and is always positive — Type ('debit' or 'credit') determines direction.
// Status is one of: posted (default), pending, reconciled, void.
type Transaction struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Amount      int    `json:"amount"` // cents, always positive
	Type        string `json:"type"`   // debit | credit
	Category    string `json:"category"`
	Account     string `json:"account"`
	Date        string `json:"date"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
	CreatedAt   string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "ledger.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS transactions(
		id TEXT PRIMARY KEY,
		description TEXT NOT NULL,
		amount INTEGER DEFAULT 0,
		type TEXT DEFAULT 'debit',
		category TEXT DEFAULT '',
		account TEXT DEFAULT '',
		date TEXT DEFAULT '',
		status TEXT DEFAULT 'posted',
		notes TEXT DEFAULT '',
		created_at TEXT DEFAULT(datetime('now'))
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tx_type ON transactions(type)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tx_category ON transactions(category)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tx_account ON transactions(account)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_tx_date ON transactions(date)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(
		resource TEXT NOT NULL,
		record_id TEXT NOT NULL,
		data TEXT NOT NULL DEFAULT '{}',
		PRIMARY KEY(resource, record_id)
	)`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string   { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) Create(e *Transaction) error {
	e.ID = genID()
	e.CreatedAt = now()
	if e.Type == "" {
		e.Type = "debit"
	}
	if e.Status == "" {
		e.Status = "posted"
	}
	if e.Date == "" {
		e.Date = time.Now().Format("2006-01-02")
	}
	_, err := d.db.Exec(
		`INSERT INTO transactions(id, description, amount, type, category, account, date, status, notes, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Description, e.Amount, e.Type, e.Category, e.Account, e.Date, e.Status, e.Notes, e.CreatedAt,
	)
	return err
}

func (d *DB) Get(id string) *Transaction {
	var e Transaction
	err := d.db.QueryRow(
		`SELECT id, description, amount, type, category, account, date, status, notes, created_at
		 FROM transactions WHERE id=?`,
		id,
	).Scan(&e.ID, &e.Description, &e.Amount, &e.Type, &e.Category, &e.Account, &e.Date, &e.Status, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Transaction {
	rows, _ := d.db.Query(
		`SELECT id, description, amount, type, category, account, date, status, notes, created_at
		 FROM transactions ORDER BY date DESC, created_at DESC`,
	)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Transaction
	for rows.Next() {
		var e Transaction
		rows.Scan(&e.ID, &e.Description, &e.Amount, &e.Type, &e.Category, &e.Account, &e.Date, &e.Status, &e.Notes, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

func (d *DB) Update(e *Transaction) error {
	_, err := d.db.Exec(
		`UPDATE transactions SET description=?, amount=?, type=?, category=?, account=?, date=?, status=?, notes=?
		 WHERE id=?`,
		e.Description, e.Amount, e.Type, e.Category, e.Account, e.Date, e.Status, e.Notes, e.ID,
	)
	return err
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM transactions WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&n)
	return n
}

func (d *DB) Search(q string, filters map[string]string) []Transaction {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (description LIKE ? OR notes LIKE ? OR account LIKE ?)"
		s := "%" + q + "%"
		args = append(args, s, s, s)
	}
	if v, ok := filters["type"]; ok && v != "" {
		where += " AND type=?"
		args = append(args, v)
	}
	if v, ok := filters["category"]; ok && v != "" {
		where += " AND category=?"
		args = append(args, v)
	}
	if v, ok := filters["status"]; ok && v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	if v, ok := filters["account"]; ok && v != "" {
		where += " AND account=?"
		args = append(args, v)
	}
	rows, _ := d.db.Query(
		`SELECT id, description, amount, type, category, account, date, status, notes, created_at
		 FROM transactions WHERE `+where+`
		 ORDER BY date DESC, created_at DESC`,
		args...,
	)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Transaction
	for rows.Next() {
		var e Transaction
		rows.Scan(&e.ID, &e.Description, &e.Amount, &e.Type, &e.Category, &e.Account, &e.Date, &e.Status, &e.Notes, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

// Stats returns financial aggregates: total tx count, total credits and
// debits in cents, net balance (credits - debits, can be negative), counts
// by status, and breakdowns by category and account.
func (d *DB) Stats() map[string]any {
	m := map[string]any{
		"total":         d.Count(),
		"total_credits": 0,
		"total_debits":  0,
		"balance":       0,
		"by_status":     map[string]int{},
		"by_category":   map[string]int{},
		"by_account":    map[string]int{},
	}

	var credits, debits int
	d.db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE type='credit' AND status != 'void'`).Scan(&credits)
	d.db.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE type='debit' AND status != 'void'`).Scan(&debits)
	m["total_credits"] = credits
	m["total_debits"] = debits
	m["balance"] = credits - debits

	if rows, _ := d.db.Query(`SELECT status, COUNT(*) FROM transactions GROUP BY status`); rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_status"] = by
	}

	if rows, _ := d.db.Query(`SELECT category, COUNT(*) FROM transactions WHERE category != '' GROUP BY category`); rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_category"] = by
	}

	if rows, _ := d.db.Query(`SELECT account, COUNT(*) FROM transactions WHERE account != '' GROUP BY account`); rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_account"] = by
	}

	return m
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
