package api

import (
	"log"
	"net/http"

	"github.com/oolong-sh/oolongd/internal/config"
)

// spawn the oolong api server
func SpawnServer() {
	mux := http.NewServeMux()

	// TODO: add some sort of JWT system for better security

	// graph endpoints
	mux.HandleFunc("GET /graph", handleGetGraph)
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// config endpoints
	mux.HandleFunc("GET /config", handleGetConfig)
	mux.HandleFunc("GET /config/graph", handleGetGraphConfig)
	mux.HandleFunc("GET /config/note-dirs", handleGetNoteDirsConfig)
	mux.HandleFunc("GET /config/default-graph-mode", handleGetGraphView)

	// note endpoints
	mux.HandleFunc("GET /notes", handleGetNotes)
	mux.HandleFunc("GET /note", handleGetNote)
	mux.HandleFunc("POST /note", handleCreateNote)
	mux.HandleFunc("PUT /note", handleUpdateNote)
	mux.HandleFunc("DELETE /note", handleDeleteNote)
	mux.HandleFunc("GET /open/note", handleOpenNote)

	// pinning endpoints
	if config.PinningEnabled() {
		mux.HandleFunc("GET /pins", handleGetPinnedNotes)
		mux.HandleFunc("POST /pins", handleAddPinnedNote)
		mux.HandleFunc("DELETE /pins", handleDeletePinnedNote)
	}

	// search endpoints
	mux.HandleFunc("GET /search/keyword", handleSearchKeyword)
	mux.HandleFunc("GET /search/note", handleSearchNote)

	// start server
	log.Println("Starting server on :11975...")
	if err := http.ListenAndServe(":11975", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
