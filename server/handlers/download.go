package handlers

import (
	"fmt"
	"log"
	"net/http"

	"filestore/server/db"
	"filestore/server/storage"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")

	log.Printf("Download request for file: %s", filename)
	if filename == "" {
		http.Error(w, "filename parameter is required", http.StatusBadRequest)
		return
	}

	servers, ok := db.FileStorageMapping.GetServers(filename)

	if !ok {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	log.Printf("File %s found in mapping: %v, servers: %v", filename, ok, servers)

	// для файла 10гб не лучшее решение, но для тестового задания - ок
	for i, storageServer := range servers {
		chunkData, err := storage.GetChunk(storageServer, filename, i)
		if err != nil {
			http.Error(w, fmt.Sprintf("error retrieving chunk %d: %v", i, err), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(chunkData)
		if err != nil {
			return
		}
	}
}
