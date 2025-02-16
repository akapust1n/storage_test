package handlers

import (
	"log"
	"sync"

	"filestore/server/db"
	"filestore/server/storage"
)

// могут быть проблемы с консистентностью данных
func doCleanup(filename string, chunks []struct {
	server string
	chunk  int
}) {
	log.Printf("Starting cleanup for interrupted upload of file: %s", filename)

	db.FileStorageMapping.Delete(filename)

	log.Printf("Cleaning up database entries for %s", filename)
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction for cleanup: %v", err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM file_chunks WHERE filename = ?", filename)
	if err != nil {
		log.Printf("Error cleaning up database entries for %s: %v", filename, err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing cleanup transaction: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, chunk := range chunks {
		wg.Add(1)
		go func(chunk struct {
			server string
			chunk  int
		}) {
			defer wg.Done()
			err := storage.DeleteChunk(chunk.server, filename, chunk.chunk)
			if err != nil {
				log.Printf("Error deleting chunk %d: %v", chunk.chunk, err)
				return
			}
			log.Printf("Successfully deleted chunk %d from %s", chunk.chunk, chunk.server)
		}(chunk)
	}
	wg.Wait()

	log.Printf("Cleanup completed for %s", filename)
}
