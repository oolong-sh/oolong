package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/oolong-sh/oolong/internal/config"
	"github.com/oolong-sh/oolong/internal/db"
	"github.com/oolong-sh/oolong/internal/state"
	"go.etcd.io/bbolt"
)

type createUpdateRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// 'GET /notes' endpoint handler returns all available note paths
func handleGetNotes(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	s := state.State()
	data := []string{}
	for k := range s.Documents {
		data = append(data, k)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// 'GET /note?path=<path>' endpoint handler gets note contents corresponding to input path
func handleGetNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter not specified ", http.StatusBadRequest)
		return
	}

	// read file contents
	b, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not read file '"+path+"'", 400)
		return
	}

	// write file contents into response body
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(string(b)); err != nil {
		log.Println(err)
		http.Error(w, "Failed to encode file contents as JSON.\n", 500)
	}
}

// 'POST /note' endpoint handler creates a note file (and any missing directories) corresponding to input path
// Expected request body: { "path": "/path/to/note", "content", "full note contents to write" }
func handleCreateNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	// parse request body
	var req createUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", 400)
	}
	log.Println("Request body: ", req)

	// check if path before file exists, then check if file exists
	if e, err := exists(req.Path); err != nil {
		log.Println(err)
		http.Error(w, "Error checking path", 500)
		return
	} else if e {
		log.Printf("File %s already exists.\n", req.Path)
		http.Error(w, "Note file already exists", 500)
		return
	}

	// create directories and file
	dir := filepath.Dir(req.Path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Println(err)
		http.Error(w, "Failed to create missing directories", 500)
		return
	}
	if err := os.WriteFile(req.Path, []byte(req.Content), 0644); err != nil {
		log.Println(err)
		http.Error(w, "Failed to create file directories", 500)
		return
	}
}

// 'PUT /note' endpoint handler updates note contents corresponding to input path
// It will create files that do not exist, but will not create directories
// Expected request body: { "path": "/path/to/note", "content", "full note contents to write" }
func handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	// parse request body
	var req createUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", 400)
	}
	log.Println("Request body: ", req)

	// write contents to file
	if err := os.WriteFile(req.Path, []byte(req.Content), 0666); err != nil {
		log.Println(err)
		http.Error(w, "Failed to write to note file", 500)
		return
	}
}

// 'Delete /note?path=/path/to/note' endpoint handler deletes a note file based on query input
func handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter not specified ", http.StatusBadRequest)
		return
	}

	// attempt to remove file
	if err := os.Remove(path); err != nil {
		log.Println("Failed to delete file", err)
		http.Error(w, "Failed to remove file", 500)
		return
	}

	// NOTE: this function may need to call the update function due to files no longer existing
	// - check this case in state, this may require substantial logic missing there
}

func addPinnedNote(path string) error {
	return db.Database.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(db.PinnedBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", db.PinnedBucket)
		}

		if bucket.Get([]byte(path)) != nil {
			return fmt.Errorf("note already pinned")
		}

		return bucket.Put([]byte(path), []byte{})
	})
}

func getPinnedNotes() ([]string, error) {
	notes := []string{}

	err := db.Database.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(db.PinnedBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", db.PinnedBucket)
		}

		return bucket.ForEach(func(k, _ []byte) error {
			notes = append(notes, string(k))
			return nil
		})
	})

	return notes, err
}

func deletePinnedNote(path string) error {
	return db.Database.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(db.PinnedBucket))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", db.PinnedBucket)
		}

		return bucket.Delete([]byte(path))
	})
}

func handleGetPinnedNotes(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	pinnedNotes, err := getPinnedNotes()
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch pinned notes", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pinnedNotes); err != nil {
		log.Println(err)
		http.Error(w, "Failed to encode response", 500)
	}
}

func handleAddPinnedNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter not specified", http.StatusBadRequest)
		return
	}

	if err := addPinnedNote(path); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleDeletePinnedNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	// CORS handling
	if err := checkOrigin(w, r); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err), 500)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter not specified", http.StatusBadRequest)
		return
	}

	if err := deletePinnedNote(path); err != nil {
		log.Println(err)
		http.Error(w, "Failed to delete pinned note", 400)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// 'GET /open/note?path=/path/to/note.md' opens the specified not file using the command specified in oolong.toml
func handleOpenNote(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method, r.URL, r.Host)

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter not specified ", http.StatusBadRequest)
		return
	}

	if e, err := exists(path); err != nil {
		log.Println(err)
		http.Error(w, "Error checking path", 500)
		return
	} else if !e {
		log.Printf("File %s does not exist.\n", path)
		http.Error(w, "Note file not found exists", 400)
		return
	}

	// open file in editor (use command defined in config)
	openCommand := append(config.OpenCommand(), path)
	cmd := exec.Command(openCommand[0], openCommand[1:]...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	if err := cmd.Run(); err != nil {
		log.Println(err)
		http.Error(w, "Error opening file in editor.", 500)
	}
}
