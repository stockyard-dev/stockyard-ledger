package server

import "github.com/stockyard-dev/stockyard-ledger/internal/license"

type Limits struct {
	MaxClients  int
	MaxInvoices int
	Recurring   bool
	PDFExport   bool
}

var freeLimits = Limits{
	MaxClients:  5,
	MaxInvoices: 20,
	Recurring:   false,
	PDFExport:   false,
}

var proLimits = Limits{
	MaxClients:  0,
	MaxInvoices: 0,
	Recurring:   true,
	PDFExport:   true,
}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}
