package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func storeHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	chunkStr := r.URL.Query().Get("chunk")
	if filename == "" || chunkStr == "" {
		http.Error(w, "filename and chunk parameters are required", http.StatusBadRequest)
		return
	}
	chunkIndex, err := strconv.Atoi(chunkStr)
	if err != nil {
		http.Error(w, "invalid chunk index", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading body", http.StatusInternalServerError)
		return
	}

	err = os.MkdirAll("data", 0755)
	if err != nil {
		http.Error(w, "error creating data directory", http.StatusInternalServerError)
		return
	}

	filePath := fmt.Sprintf("data/%s_%d", filename, chunkIndex)
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		http.Error(w, "error saving file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stored"))
}

func retrieveHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	chunkStr := r.URL.Query().Get("chunk")
	if filename == "" || chunkStr == "" {
		http.Error(w, "filename and chunk parameters are required", http.StatusBadRequest)
		return
	}
	chunkIndex, err := strconv.Atoi(chunkStr)
	if err != nil {
		http.Error(w, "invalid chunk index", http.StatusBadRequest)
		return
	}
	filePath := fmt.Sprintf("data/%s_%d", filename, chunkIndex)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "error reading file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	chunkStr := r.URL.Query().Get("chunk")

	if filename == "" || chunkStr == "" {
		http.Error(w, "filename and chunk parameters are required", http.StatusBadRequest)
		return
	}

	chunkPath := filepath.Join("data", filename+"_"+chunkStr)

	if err := os.Remove(chunkPath); err != nil && !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("error deleting chunk: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/retrieve", retrieveHandler)
	http.HandleFunc("/delete", deleteHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Storage server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
