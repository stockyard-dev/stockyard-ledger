// Stockyard Ledger — Invoicing for freelancers.
// Create clients, send invoices, track payments. Self-hosted.
// Single binary, embedded SQLite, zero external dependencies.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/stockyard-dev/stockyard-ledger/internal/license"
	"github.com/stockyard-dev/stockyard-ledger/internal/server"
	"github.com/stockyard-dev/stockyard-ledger/internal/store"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Printf("ledger %s\n", version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && (os.Args[1] == "--health" || os.Args[1] == "health") {
		fmt.Println("ok")
		os.Exit(0)
	}

	log.SetFlags(log.Ltime | log.Lshortfile)

	port := 8920
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	licenseKey := os.Getenv("LEDGER_LICENSE_KEY")
	licInfo, licErr := license.Validate(licenseKey, "ledger")
	if licenseKey != "" && licErr != nil {
		log.Printf("[license] WARNING: %v — running in free tier", licErr)
		licInfo = nil
	}
	limits := server.LimitsFor(licInfo)
	if licInfo != nil && licInfo.IsPro() {
		log.Printf("  License:   Pro (%s)", licInfo.CustomerID)
	} else {
		log.Printf("  License:   Free tier (set LEDGER_LICENSE_KEY to unlock Pro)")
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	log.Printf("")
	log.Printf("  Stockyard Ledger %s", version)
	log.Printf("  API:            http://localhost:%d/api/invoices", port)
	log.Printf("  Invoice view:   http://localhost:%d/invoice/{id}", port)
	log.Printf("  Dashboard:      http://localhost:%d/ui", port)
	log.Printf("")

	srv := server.New(db, port, limits)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
