package handlers

import (
	"log"
	"net/http"

	"filestore/server/storage"
)

// возможно, переусложнил требование по добавлению новых серверов
func RegisterStorageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serverURL := r.URL.Query().Get("url")
	if serverURL == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	storage.DefaultManager.AddServer(serverURL)

	log.Printf("New storage server registered: %s", serverURL)
	w.WriteHeader(http.StatusOK)
}
