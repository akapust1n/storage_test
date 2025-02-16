package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	dbpkg "filestore/server/db"
	"filestore/server/handlers"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() error {
	dbHost := getEnvOrDefault("MYSQL_HOST", "localhost")
	dbUser := getEnvOrDefault("MYSQL_USER", "root")
	dbPass := getEnvOrDefault("MYSQL_PASSWORD", "password")
	dbName := getEnvOrDefault("MYSQL_DATABASE", "filestore")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", dbUser, dbPass, dbHost, dbName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	dbpkg.DB = db

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS file_chunks (
			filename VARCHAR(255),
			server_url TEXT,
			chunk_order INT,
			PRIMARY KEY (filename, chunk_order)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return dbpkg.LoadMappings()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/upload", handlers.UploadHandler)
	http.HandleFunc("/download", handlers.DownloadHandler)
	http.HandleFunc("/register", handlers.RegisterStorageHandler)

	port := getEnvOrDefault("PORT", "8080")
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
