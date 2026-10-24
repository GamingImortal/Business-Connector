// Command backend serves the Business Connect REST API, storing data in SQLite.
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	dbPath := flag.String("db", "business-connect.db", "SQLite database file")
	origins := flag.String("origins", "http://localhost:5173,http://localhost:5174", "comma-separated origins allowed by CORS")
	flag.Parse()

	store, err := openStore(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Seed(); err != nil {
		log.Fatal(err)
	}
	if err := store.EnsureConcierge(); err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           NewServer(store, strings.Split(*origins, ",")).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("Business Connect API listening on %s (db: %s)", *addr, *dbPath)
	log.Fatal(srv.ListenAndServe())
}
