# Stockyard Ledger

**Invoicing for freelancers.** Create clients, send invoices with line items, track payment status. Every freelancer pays $15/mo for FreshBooks to do something this simple. Single binary, no external dependencies.

Part of the [Stockyard](https://stockyard.dev) suite of self-hosted developer tools.

## Quick Start

```bash
curl -sfL https://stockyard.dev/install/ledger | sh
ledger
```

## Usage

```bash
# Create a client
curl -X POST http://localhost:8920/api/clients \
  -H "Content-Type: application/json" \
  -d '{"name":"Acme Inc","email":"billing@acme.com"}'

# Create an invoice with line items
curl -X POST http://localhost:8920/api/invoices \
  -H "Content-Type: application/json" \
  -d '{"client_id":"{id}","due_date":"2026-05-01","items":[
    {"description":"Website redesign","quantity":1,"unit_price_cents":250000},
    {"description":"Consulting (10 hrs)","quantity":10,"unit_price_cents":15000}
  ]}'

# Share invoice link with client
open http://localhost:8920/invoice/{invoice_id}

# Mark as paid
curl -X PUT http://localhost:8920/api/invoices/{id}/status \
  -H "Content-Type: application/json" \
  -d '{"status":"paid"}'
```

## Free vs Pro

| Feature | Free | Pro ($2.99/mo) |
|---------|------|----------------|
| Clients | 5 | Unlimited |
| Invoices | 20 | Unlimited |
| Public invoice view | ✓ | ✓ |
| Auto-numbering | ✓ | ✓ |
| Recurring invoices | — | ✓ |
| PDF export | — | ✓ |

## License

Apache 2.0 — see [LICENSE](LICENSE).
