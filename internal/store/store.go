package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Account struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Currency string `json:"currency"`
	Balance float64 `json:"balance"`
	Description string `json:"description"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"ledger.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS accounts(id TEXT PRIMARY KEY,name TEXT NOT NULL,type TEXT DEFAULT 'asset',currency TEXT DEFAULT 'USD',balance REAL DEFAULT 0,description TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Account)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO accounts(id,name,type,currency,balance,description,created_at)VALUES(?,?,?,?,?,?,?)`,e.ID,e.Name,e.Type,e.Currency,e.Balance,e.Description,e.CreatedAt);return err}
func(d *DB)Get(id string)*Account{var e Account;if d.db.QueryRow(`SELECT id,name,type,currency,balance,description,created_at FROM accounts WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.Type,&e.Currency,&e.Balance,&e.Description,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Account{rows,_:=d.db.Query(`SELECT id,name,type,currency,balance,description,created_at FROM accounts ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Account;for rows.Next(){var e Account;rows.Scan(&e.ID,&e.Name,&e.Type,&e.Currency,&e.Balance,&e.Description,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM accounts WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM accounts`).Scan(&n);return n}
