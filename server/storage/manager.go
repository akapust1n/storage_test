package storage

import (
	"os"
	"strings"
	"sync"
)

type Manager struct {
	sync.RWMutex
	servers []string
}

var DefaultManager = &Manager{}

func (m *Manager) GetServers() []string {
	m.RLock()
	defer m.RUnlock()

	if len(m.servers) > 0 {
		return m.servers
	}

	env := os.Getenv("STORAGE_SERVERS")
	if env == "" {
		return []string{
			"http://localhost:8081",
			"http://localhost:8082",
			"http://localhost:8083",
			"http://localhost:8084",
			"http://localhost:8085",
			"http://localhost:8086",
			"http://localhost:8087",
		}
	}

	parts := strings.Split(env, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

func (m *Manager) AddServer(url string) {
	m.Lock()
	m.servers = append(m.servers, url)
	m.Unlock()
}
