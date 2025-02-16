package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"filestore/server/db"
	"filestore/server/storage"
)

// вся эта функция имеет проблемы с консистентностью
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filename := r.URL.Query().Get("filename")
	log.Printf("Starting upload for file: %s", filename)
	if filename == "" {
		http.Error(w, "filename parameter is required", http.StatusBadRequest)
		return
	}

	db.FileStorageMapping.Delete(filename)

	var uploadedChunks []struct {
		server string
		chunk  int
	}

	servers := storage.DefaultManager.GetServers()
	if len(servers) == 0 {
		http.Error(w, "no storage servers available", http.StatusInternalServerError)
		return
	}

	contentLengthStr := r.Header.Get("Content-Length")
	if contentLengthStr == "" {
		http.Error(w, "Content-Length header is required", http.StatusBadRequest)
		return
	}
	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Content-Length", http.StatusBadRequest)
		return
	}

	chunkCount := len(servers)
	chunkSize := contentLength / int64(chunkCount)
	remainder := contentLength % int64(chunkCount)

	var usedServers []string
	var tempMapping []string

	for i := 0; i < chunkCount; i++ {
		storageServer := servers[i]

		select {
		case <-ctx.Done():
			log.Printf("Upload interrupted for file %s, starting cleanup", filename)
			doCleanup(filename, uploadedChunks)
			log.Printf("Cleanup completed for interrupted upload of %s", filename)
			return
		default:
		}

		currentChunkSize := chunkSize
		// в последнем сервере будет чуть больше данных
		if i == chunkCount-1 {
			currentChunkSize += remainder
		}

		limitedReader := io.LimitReader(r.Body, currentChunkSize)
		chunkData, err := ioutil.ReadAll(limitedReader)
		if err != nil {
			log.Printf("Error reading chunk for file %s: %v", filename, err)
			doCleanup(filename, uploadedChunks)
			http.Error(w, "error reading file chunk", http.StatusInternalServerError)
			return
		}

		// код может быть красивее :)
		select {
		case <-ctx.Done():
			log.Printf("Upload interrupted for file %s after reading chunk", filename)
			doCleanup(filename, uploadedChunks)
			log.Printf("Cleanup completed for interrupted upload of %s", filename)
			return
		default:
		}

		err = storage.SendChunk(storageServer, filename, i, chunkData)
		if err != nil {
			doCleanup(filename, uploadedChunks)
			http.Error(w, fmt.Sprintf("error storing chunk %d: %v", i, err), http.StatusInternalServerError)
			return
		}

		select {
		case <-ctx.Done():
			log.Printf("Upload interrupted for file %s after storing chunk", filename)
			doCleanup(filename, uploadedChunks)
			log.Printf("Cleanup completed for interrupted upload of %s", filename)
			return
		default:
		}

		uploadedChunks = append(uploadedChunks, struct {
			server string
			chunk  int
		}{storageServer, i})
		tempMapping = append(tempMapping, storageServer)
	}

	tx, err := db.DB.Begin()
	if err != nil {
		doCleanup(filename, uploadedChunks)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM file_chunks WHERE filename = ?", filename)
	if err != nil {
		doCleanup(filename, uploadedChunks)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	select {
	case <-ctx.Done():
		log.Printf("Upload interrupted for file %s before DB commit", filename)
		doCleanup(filename, uploadedChunks)
		log.Printf("Cleanup completed for interrupted upload of %s", filename)
		return
	default:
	}

	// возможно, переусложнил в плане персистентности данных и можно было без бд.
	for i, server := range tempMapping {
		_, err = tx.Exec("INSERT INTO file_chunks (filename, server_url, chunk_order) VALUES (?, ?, ?)",
			filename, server, i)
		if err != nil {
			doCleanup(filename, uploadedChunks)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		usedServers = append(usedServers, server)
	}

	if err := tx.Commit(); err != nil {
		doCleanup(filename, uploadedChunks)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	db.FileStorageMapping.Set(filename, usedServers)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Upload successful"))
}
