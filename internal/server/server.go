package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/stockyard-dev/stockyard-ledger/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/clients", s.handleCreateClient)
	s.mux.HandleFunc("GET /api/clients", s.handleListClients)
	s.mux.HandleFunc("GET /api/clients/{id}", s.handleGetClient)
	s.mux.HandleFunc("DELETE /api/clients/{id}", s.handleDeleteClient)

	s.mux.HandleFunc("POST /api/invoices", s.handleCreateInvoice)
	s.mux.HandleFunc("GET /api/invoices", s.handleListInvoices)
	s.mux.HandleFunc("GET /api/invoices/{id}", s.handleGetInvoice)
	s.mux.HandleFunc("PUT /api/invoices/{id}/status", s.handleUpdateStatus)
	s.mux.HandleFunc("DELETE /api/invoices/{id}", s.handleDeleteInvoice)

	// Public invoice view
	s.mux.HandleFunc("GET /invoice/{id}", s.handlePublicInvoice)

	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-ledger", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[ledger] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// --- Client handlers ---

func (s *Server) handleCreateClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Company string `json:"company"`
		Address string `json:"address"`
		Notes   string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "name is required"})
		return
	}
	if s.limits.MaxClients > 0 {
		clients, _ := s.db.ListClients()
		if LimitReached(s.limits.MaxClients, len(clients)) {
			writeJSON(w, 402, map[string]string{"error": fmt.Sprintf("free tier limit: %d clients — upgrade to Pro", s.limits.MaxClients), "upgrade": "https://stockyard.dev/ledger/"})
			return
		}
	}
	c, err := s.db.CreateClient(req.Name, req.Email, req.Company, req.Address, req.Notes)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"client": c})
}

func (s *Server) handleListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := s.db.ListClients()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if clients == nil {
		clients = []store.Client{}
	}
	writeJSON(w, 200, map[string]any{"clients": clients, "count": len(clients)})
}

func (s *Server) handleGetClient(w http.ResponseWriter, r *http.Request) {
	c, err := s.db.GetClient(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "client not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"client": c})
}

func (s *Server) handleDeleteClient(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteClient(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Invoice handlers ---

func (s *Server) handleCreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID  string           `json:"client_id"`
		IssueDate string           `json:"issue_date"`
		DueDate   string           `json:"due_date"`
		Currency  string           `json:"currency"`
		Notes     string           `json:"notes"`
		Items     []store.LineItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.ClientID == "" {
		writeJSON(w, 400, map[string]string{"error": "client_id is required"})
		return
	}
	if _, err := s.db.GetClient(req.ClientID); err != nil {
		writeJSON(w, 404, map[string]string{"error": "client not found"})
		return
	}
	if s.limits.MaxInvoices > 0 {
		total := s.db.TotalInvoices()
		if LimitReached(s.limits.MaxInvoices, total) {
			writeJSON(w, 402, map[string]string{"error": fmt.Sprintf("free tier limit: %d invoices — upgrade to Pro", s.limits.MaxInvoices), "upgrade": "https://stockyard.dev/ledger/"})
			return
		}
	}
	inv, err := s.db.CreateInvoice(req.ClientID, req.IssueDate, req.DueDate, req.Currency, req.Notes, req.Items)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	viewURL := fmt.Sprintf("http://localhost:%d/invoice/%s", s.port, inv.ID)
	writeJSON(w, 201, map[string]any{"invoice": inv, "view_url": viewURL})
}

func (s *Server) handleListInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := s.db.ListInvoices()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if invoices == nil {
		invoices = []store.Invoice{}
	}
	writeJSON(w, 200, map[string]any{"invoices": invoices, "count": len(invoices)})
}

func (s *Server) handleGetInvoice(w http.ResponseWriter, r *http.Request) {
	inv, err := s.db.GetInvoice(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "invoice not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"invoice": inv})
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Status string `json:"status"` // draft, sent, paid, overdue, cancelled
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Status == "" {
		writeJSON(w, 400, map[string]string{"error": "status is required (draft, sent, paid, overdue, cancelled)"})
		return
	}
	s.db.UpdateInvoiceStatus(id, req.Status)
	inv, _ := s.db.GetInvoice(id)
	writeJSON(w, 200, map[string]any{"invoice": inv})
}

func (s *Server) handleDeleteInvoice(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteInvoice(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Public invoice view ---

func (s *Server) handlePublicInvoice(w http.ResponseWriter, r *http.Request) {
	inv, err := s.db.GetInvoice(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	client, _ := s.db.GetClient(inv.ClientID)
	bizName := s.db.GetSetting("business_name")

	var itemsHTML strings.Builder
	for _, li := range inv.Items {
		itemsHTML.WriteString(fmt.Sprintf(`<tr><td style="padding:.6rem .8rem;border-bottom:1px solid #2e261e;color:#bfb5a3">%s</td>
			<td style="padding:.6rem .8rem;border-bottom:1px solid #2e261e;color:#bfb5a3;text-align:right">%.2f</td>
			<td style="padding:.6rem .8rem;border-bottom:1px solid #2e261e;color:#bfb5a3;text-align:right">$%.2f</td>
			<td style="padding:.6rem .8rem;border-bottom:1px solid #2e261e;color:#f0e6d3;text-align:right;font-weight:600">$%.2f</td></tr>`,
			he(li.Description), li.Quantity, float64(li.UnitPriceCents)/100, float64(li.TotalCents)/100))
	}

	clientName := ""
	if client != nil {
		clientName = client.Name
	}

	statusColor := "#d4a843"
	switch inv.Status {
	case "paid":
		statusColor = "#5ba86e"
	case "overdue":
		statusColor = "#c0392b"
	case "sent":
		statusColor = "#4a90d9"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, invoiceTemplate, he(inv.Number), he(bizName), he(inv.Number), statusColor, statusColor, strings.ToUpper(inv.Status),
		he(clientName), inv.IssueDate, inv.DueDate, itemsHTML.String(),
		float64(inv.TotalCents)/100, inv.Currency)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func he(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

const invoiceTemplate = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Invoice %s</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&family=Libre+Baskerville:wght@400;700&display=swap" rel="stylesheet">
<style>body{background:#1a1410;color:#f0e6d3;font-family:'Libre Baskerville',serif;margin:0;min-height:100vh;padding:2rem}
.inv{max-width:700px;margin:0 auto;background:#241e18;border:1px solid #2e261e;padding:2.5rem}
.top{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:2rem}
.from{font-family:'JetBrains Mono',monospace;font-size:.78rem;color:#a0845c}
.inv-num{font-family:'JetBrains Mono',monospace;font-size:1.4rem;font-weight:700;color:#f0e6d3}
.badge{font-family:'JetBrains Mono',monospace;font-size:.6rem;padding:.2rem .6rem;letter-spacing:1px;border:1px solid;display:inline-block;margin-top:.3rem}
.meta{display:grid;grid-template-columns:1fr 1fr;gap:1.5rem;margin-bottom:2rem;font-family:'JetBrains Mono',monospace;font-size:.75rem}
.meta-lbl{font-size:.6rem;letter-spacing:2px;text-transform:uppercase;color:#7a7060;margin-bottom:.2rem}
table{width:100%%;border-collapse:collapse;font-family:'JetBrains Mono',monospace;font-size:.78rem}
th{background:#2e261e;padding:.5rem .8rem;text-align:left;color:#c4a87a;font-weight:400;font-size:.62rem;letter-spacing:1px;text-transform:uppercase}
.total-row{border-top:2px solid #c45d2c;font-weight:700;font-size:1rem;color:#f0e6d3}
.footer{text-align:center;margin-top:2rem;font-size:.55rem;color:#7a7060;font-family:'JetBrains Mono',monospace}
.footer a{color:#e8753a;text-decoration:none}
</style></head><body>
<div class="inv">
<div class="top">
<div><div class="from">%s</div></div>
<div style="text-align:right"><div class="inv-num">%s</div><div class="badge" style="color:%s;border-color:%s">%s</div></div>
</div>
<div class="meta">
<div><div class="meta-lbl">Bill To</div><div style="color:#f0e6d3">%s</div></div>
<div style="text-align:right"><div class="meta-lbl">Issue Date</div><div>%s</div><div class="meta-lbl" style="margin-top:.5rem">Due Date</div><div>%s</div></div>
</div>
<table>
<thead><tr><th>Description</th><th style="text-align:right">Qty</th><th style="text-align:right">Price</th><th style="text-align:right">Total</th></tr></thead>
<tbody>%s</tbody>
<tfoot><tr class="total-row"><td colspan="3" style="padding:.8rem;text-align:right">Total</td><td style="padding:.8rem;text-align:right">$%.2f %s</td></tr></tfoot>
</table>
</div>
<div class="footer">Powered by <a href="https://stockyard.dev/ledger/">Stockyard Ledger</a></div>
</body></html>`
