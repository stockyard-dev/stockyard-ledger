package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "ledger.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT DEFAULT '',
    company TEXT DEFAULT '',
    address TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS invoices (
    id TEXT PRIMARY KEY,
    number TEXT NOT NULL UNIQUE,
    client_id TEXT NOT NULL,
    status TEXT DEFAULT 'draft',
    issue_date TEXT DEFAULT '',
    due_date TEXT DEFAULT '',
    subtotal_cents INTEGER DEFAULT 0,
    tax_cents INTEGER DEFAULT 0,
    total_cents INTEGER DEFAULT 0,
    currency TEXT DEFAULT 'USD',
    notes TEXT DEFAULT '',
    paid_at TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_inv_client ON invoices(client_id);
CREATE INDEX IF NOT EXISTS idx_inv_status ON invoices(status);

CREATE TABLE IF NOT EXISTS line_items (
    id TEXT PRIMARY KEY,
    invoice_id TEXT NOT NULL,
    description TEXT NOT NULL,
    quantity REAL DEFAULT 1,
    unit_price_cents INTEGER DEFAULT 0,
    total_cents INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_items_inv ON line_items(invoice_id);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT DEFAULT ''
);
`)
	db.conn.Exec("INSERT OR IGNORE INTO settings (key,value) VALUES ('business_name','')")
	db.conn.Exec("INSERT OR IGNORE INTO settings (key,value) VALUES ('business_email','')")
	db.conn.Exec("INSERT OR IGNORE INTO settings (key,value) VALUES ('business_address','')")
	db.conn.Exec("INSERT OR IGNORE INTO settings (key,value) VALUES ('next_invoice_number','1001')")
	return err
}

// --- Clients ---

type Client struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Company   string `json:"company"`
	Address   string `json:"address"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) CreateClient(name, email, company, address, notes string) (*Client, error) {
	id := "cli_" + genID(6)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO clients (id,name,email,company,address,notes,created_at) VALUES (?,?,?,?,?,?,?)",
		id, name, email, company, address, notes, now)
	if err != nil {
		return nil, err
	}
	return &Client{ID: id, Name: name, Email: email, Company: company, Address: address, Notes: notes, CreatedAt: now}, nil
}

func (db *DB) ListClients() ([]Client, error) {
	rows, err := db.conn.Query("SELECT id,name,email,company,address,notes,created_at FROM clients ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Client
	for rows.Next() {
		var c Client
		rows.Scan(&c.ID, &c.Name, &c.Email, &c.Company, &c.Address, &c.Notes, &c.CreatedAt)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (db *DB) GetClient(id string) (*Client, error) {
	var c Client
	err := db.conn.QueryRow("SELECT id,name,email,company,address,notes,created_at FROM clients WHERE id=?", id).
		Scan(&c.ID, &c.Name, &c.Email, &c.Company, &c.Address, &c.Notes, &c.CreatedAt)
	return &c, err
}

func (db *DB) DeleteClient(id string) error {
	_, err := db.conn.Exec("DELETE FROM clients WHERE id=?", id)
	return err
}

// --- Invoices ---

type Invoice struct {
	ID           string     `json:"id"`
	Number       string     `json:"number"`
	ClientID     string     `json:"client_id"`
	ClientName   string     `json:"client_name,omitempty"`
	Status       string     `json:"status"`
	IssueDate    string     `json:"issue_date"`
	DueDate      string     `json:"due_date"`
	SubtotalCents int       `json:"subtotal_cents"`
	TaxCents     int        `json:"tax_cents"`
	TotalCents   int        `json:"total_cents"`
	Currency     string     `json:"currency"`
	Notes        string     `json:"notes"`
	PaidAt       string     `json:"paid_at,omitempty"`
	CreatedAt    string     `json:"created_at"`
	Items        []LineItem `json:"items,omitempty"`
}

type LineItem struct {
	ID             string  `json:"id"`
	InvoiceID      string  `json:"invoice_id"`
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"`
	UnitPriceCents int     `json:"unit_price_cents"`
	TotalCents     int     `json:"total_cents"`
}

func (db *DB) CreateInvoice(clientID, issueDate, dueDate, currency, notes string, items []LineItem) (*Invoice, error) {
	id := "inv_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)

	// Get next invoice number
	var numStr string
	db.conn.QueryRow("SELECT value FROM settings WHERE key='next_invoice_number'").Scan(&numStr)
	number := "INV-" + numStr

	// Calculate totals
	var subtotal int
	for i := range items {
		items[i].TotalCents = int(float64(items[i].UnitPriceCents) * items[i].Quantity)
		subtotal += items[i].TotalCents
	}

	if currency == "" {
		currency = "USD"
	}
	if issueDate == "" {
		issueDate = time.Now().Format("2006-01-02")
	}

	_, err := db.conn.Exec("INSERT INTO invoices (id,number,client_id,issue_date,due_date,subtotal_cents,total_cents,currency,notes,created_at) VALUES (?,?,?,?,?,?,?,?,?,?)",
		id, number, clientID, issueDate, dueDate, subtotal, subtotal, currency, notes, now)
	if err != nil {
		return nil, err
	}

	// Insert line items
	for _, item := range items {
		itemID := "li_" + genID(6)
		db.conn.Exec("INSERT INTO line_items (id,invoice_id,description,quantity,unit_price_cents,total_cents) VALUES (?,?,?,?,?,?)",
			itemID, id, item.Description, item.Quantity, item.UnitPriceCents, item.TotalCents)
	}

	// Increment invoice number
	var num int
	fmt.Sscanf(numStr, "%d", &num)
	db.conn.Exec("UPDATE settings SET value=? WHERE key='next_invoice_number'", fmt.Sprintf("%d", num+1))

	return &Invoice{ID: id, Number: number, ClientID: clientID, Status: "draft",
		IssueDate: issueDate, DueDate: dueDate, SubtotalCents: subtotal, TotalCents: subtotal,
		Currency: currency, Notes: notes, CreatedAt: now, Items: items}, nil
}

func (db *DB) ListInvoices() ([]Invoice, error) {
	rows, err := db.conn.Query(`SELECT i.id,i.number,i.client_id,COALESCE(c.name,''),i.status,i.issue_date,i.due_date,
		i.subtotal_cents,i.tax_cents,i.total_cents,i.currency,i.notes,i.paid_at,i.created_at
		FROM invoices i LEFT JOIN clients c ON c.id=i.client_id ORDER BY i.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Invoice
	for rows.Next() {
		var inv Invoice
		rows.Scan(&inv.ID, &inv.Number, &inv.ClientID, &inv.ClientName, &inv.Status, &inv.IssueDate,
			&inv.DueDate, &inv.SubtotalCents, &inv.TaxCents, &inv.TotalCents, &inv.Currency, &inv.Notes,
			&inv.PaidAt, &inv.CreatedAt)
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (db *DB) GetInvoice(id string) (*Invoice, error) {
	var inv Invoice
	err := db.conn.QueryRow(`SELECT i.id,i.number,i.client_id,COALESCE(c.name,''),i.status,i.issue_date,i.due_date,
		i.subtotal_cents,i.tax_cents,i.total_cents,i.currency,i.notes,i.paid_at,i.created_at
		FROM invoices i LEFT JOIN clients c ON c.id=i.client_id WHERE i.id=?`, id).
		Scan(&inv.ID, &inv.Number, &inv.ClientID, &inv.ClientName, &inv.Status, &inv.IssueDate,
			&inv.DueDate, &inv.SubtotalCents, &inv.TaxCents, &inv.TotalCents, &inv.Currency, &inv.Notes,
			&inv.PaidAt, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	// Load line items
	itemRows, _ := db.conn.Query("SELECT id,invoice_id,description,quantity,unit_price_cents,total_cents FROM line_items WHERE invoice_id=? ORDER BY sort_order", id)
	if itemRows != nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var li LineItem
			itemRows.Scan(&li.ID, &li.InvoiceID, &li.Description, &li.Quantity, &li.UnitPriceCents, &li.TotalCents)
			inv.Items = append(inv.Items, li)
		}
	}
	return &inv, nil
}

func (db *DB) UpdateInvoiceStatus(id, status string) {
	if status == "paid" {
		now := time.Now().UTC().Format(time.RFC3339)
		db.conn.Exec("UPDATE invoices SET status=?, paid_at=? WHERE id=?", status, now, id)
	} else {
		db.conn.Exec("UPDATE invoices SET status=? WHERE id=?", status, id)
	}
}

func (db *DB) DeleteInvoice(id string) error {
	db.conn.Exec("DELETE FROM line_items WHERE invoice_id=?", id)
	_, err := db.conn.Exec("DELETE FROM invoices WHERE id=?", id)
	return err
}

func (db *DB) TotalInvoices() int {
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM invoices").Scan(&count)
	return count
}

// --- Settings ---

func (db *DB) GetSetting(key string) string {
	var val string
	db.conn.QueryRow("SELECT value FROM settings WHERE key=?", key).Scan(&val)
	return val
}

func (db *DB) SetSetting(key, value string) {
	db.conn.Exec("INSERT OR REPLACE INTO settings (key,value) VALUES (?,?)", key, value)
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var clients, invoices, paid, outstanding, totalRevenue int
	db.conn.QueryRow("SELECT COUNT(*) FROM clients").Scan(&clients)
	db.conn.QueryRow("SELECT COUNT(*) FROM invoices").Scan(&invoices)
	db.conn.QueryRow("SELECT COUNT(*) FROM invoices WHERE status='paid'").Scan(&paid)
	db.conn.QueryRow("SELECT COUNT(*) FROM invoices WHERE status IN ('sent','overdue')").Scan(&outstanding)
	db.conn.QueryRow("SELECT COALESCE(SUM(total_cents),0) FROM invoices WHERE status='paid'").Scan(&totalRevenue)
	return map[string]any{"clients": clients, "invoices": invoices, "paid": paid,
		"outstanding": outstanding, "total_revenue_cents": totalRevenue}
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
