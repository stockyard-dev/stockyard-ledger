package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Transaction struct {
	ID string `json:"id"`
	Description string `json:"description"`
	Amount int `json:"amount"`
	Type string `json:"type"`
	Category string `json:"category"`
	Account string `json:"account"`
	Date string `json:"date"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"ledger.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS transactions(id TEXT PRIMARY KEY,description TEXT NOT NULL,amount INTEGER DEFAULT 0,type TEXT DEFAULT 'debit',category TEXT DEFAULT '',account TEXT DEFAULT '',date TEXT DEFAULT '',status TEXT DEFAULT 'posted',notes TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Transaction)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO transactions(id,description,amount,type,category,account,date,status,notes,created_at)VALUES(?,?,?,?,?,?,?,?,?,?)`,e.ID,e.Description,e.Amount,e.Type,e.Category,e.Account,e.Date,e.Status,e.Notes,e.CreatedAt);return err}
func(d *DB)Get(id string)*Transaction{var e Transaction;if d.db.QueryRow(`SELECT id,description,amount,type,category,account,date,status,notes,created_at FROM transactions WHERE id=?`,id).Scan(&e.ID,&e.Description,&e.Amount,&e.Type,&e.Category,&e.Account,&e.Date,&e.Status,&e.Notes,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Transaction{rows,_:=d.db.Query(`SELECT id,description,amount,type,category,account,date,status,notes,created_at FROM transactions ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Transaction;for rows.Next(){var e Transaction;rows.Scan(&e.ID,&e.Description,&e.Amount,&e.Type,&e.Category,&e.Account,&e.Date,&e.Status,&e.Notes,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Transaction)error{_,err:=d.db.Exec(`UPDATE transactions SET description=?,amount=?,type=?,category=?,account=?,date=?,status=?,notes=? WHERE id=?`,e.Description,e.Amount,e.Type,e.Category,e.Account,e.Date,e.Status,e.Notes,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM transactions WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Transaction{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (description LIKE ?)"
        args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["type"];ok&&v!=""{where+=" AND type=?";args=append(args,v)}
    if v,ok:=filters["category"];ok&&v!=""{where+=" AND category=?";args=append(args,v)}
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,description,amount,type,category,account,date,status,notes,created_at FROM transactions WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Transaction;for rows.Next(){var e Transaction;rows.Scan(&e.ID,&e.Description,&e.Amount,&e.Type,&e.Category,&e.Account,&e.Date,&e.Status,&e.Notes,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM transactions GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
