package db

import (
	"os"
	"path/filepath"
	"runtime"

	"go.etcd.io/bbolt"
)

var Database *bbolt.DB

const PinnedBucket = "PinnedNotes"

// Initialize the database and ensure the bucket exists
func InitializeDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var db_path string
	if runtime.GOOS == "windows" {
		db_path = filepath.Join(homeDir, "AppData", "Local", "oolong", "oolong.db")
	} else {
		db_path = filepath.Join(homeDir, ".local", "share", "oolong", "oolong.db")
	}

	err = os.MkdirAll(filepath.Dir(db_path), 0755)
	if err != nil {
		return err
	}

	Database, err = bbolt.Open(db_path, 0666, nil)
	if err != nil {
		return err
	}

	return Database.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(PinnedBucket))
		return err
	})
}

// Close the database when the server shuts down
func CloseDB() {
	if Database != nil {
		Database.Close()
	}
}
