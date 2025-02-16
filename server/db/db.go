package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type FileMapping struct {
	sync.RWMutex
	m map[string][]string
}

var DB *sql.DB
var FileStorageMapping = &FileMapping{m: make(map[string][]string)}

func Init(host, user, pass, dbname string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", user, pass, host, dbname)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}

	_, err = DB.Exec(`
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

	return LoadMappings()
}

func LoadMappings() error {
	rows, err := DB.Query("SELECT filename, GROUP_CONCAT(server_url ORDER BY chunk_order) as servers FROM file_chunks GROUP BY filename")
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer rows.Close()

	FileStorageMapping.Lock()
	defer FileStorageMapping.Unlock()

	for rows.Next() {
		var filename, serversStr string
		if err := rows.Scan(&filename, &serversStr); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		FileStorageMapping.m[filename] = strings.Split(serversStr, ",")
	}
	return rows.Err()
}

func (fm *FileMapping) Delete(filename string) {
	fm.Lock()
	delete(fm.m, filename)
	fm.Unlock()
}

func (fm *FileMapping) Set(filename string, servers []string) {
	fm.Lock()
	fm.m[filename] = servers
	fm.Unlock()
}

func (fm *FileMapping) GetServers(filename string) ([]string, bool) {
	fm.RLock()
	servers, ok := fm.m[filename]
	fm.RUnlock()
	return servers, ok
}
