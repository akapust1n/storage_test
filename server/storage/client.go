package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// не идеальное решение для больших файлов
func SendChunk(server, filename string, chunkIndex int, data []byte) error {
	url := fmt.Sprintf("%s/store?filename=%s&chunk=%d", server, filename, chunkIndex)
	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("storage server returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func GetChunk(server, filename string, chunkIndex int) ([]byte, error) {
	url := fmt.Sprintf("%s/retrieve?filename=%s&chunk=%d", server, filename, chunkIndex)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("storage server returned status %d: %s", resp.StatusCode, string(body))
	}
	return ioutil.ReadAll(resp.Body)
}

func DeleteChunk(server, filename string, chunkIndex int) error {
	url := fmt.Sprintf("%s/delete?filename=%s&chunk=%d", server, filename, chunkIndex)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
