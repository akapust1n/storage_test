package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	serverAddr      = "http://localhost:8080"
	testFileName    = "test_file.txt"
	testFileContent = "test file content 123"
	largeFileSize   = 1024 * 1024 * 100 // 100MB
)

func TestIntegrationUploadDownload(t *testing.T) {
	t.Run("Upload File", func(t *testing.T) {
		content := []byte(testFileContent)

		req, err := http.NewRequest("POST",
			fmt.Sprintf("%s/upload?filename=%s", serverAddr, testFileName),
			bytes.NewReader(content))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(content)))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to upload file: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Upload failed with status %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("Download File", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/download?filename=%s", serverAddr, testFileName))
		if err != nil {
			t.Fatalf("Failed to download file: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Download failed with status %d: %s", resp.StatusCode, string(body))
		}

		downloadedContent, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read downloaded content: %v", err)
		}

		if string(downloadedContent) != testFileContent {
			t.Errorf("Downloaded content does not match original content.\nExpected: %s\nGot: %s",
				testFileContent, string(downloadedContent))
		}
	})

	t.Run("Download Non-existent File", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/download?filename=nonexistent.txt", serverAddr))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status NotFound, got %v", resp.StatusCode)
		}
	})

	t.Run("Interrupted Upload", func(t *testing.T) {
		content := make([]byte, largeFileSize)
		for i := range content {
			content[i] = byte(i % 256)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST",
			fmt.Sprintf("%s/upload?filename=interrupted.txt", serverAddr),
			bytes.NewReader(content))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(content)))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if ctx.Err() != context.DeadlineExceeded {
				t.Fatalf("Expected context deadline exceeded error, got: %v", err)
			}
			t.Logf("Upload was interrupted as expected")
		} else {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Unexpected successful response: %s", string(body))
			resp.Body.Close()
		}

		// даем время на очистку
		time.Sleep(2 * time.Second)

		resp, err = http.Get(fmt.Sprintf("%s/download?filename=interrupted.txt", serverAddr))
		if err != nil {
			t.Fatalf("Error checking file existence: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("File is still accessible after cleanup. Status: %d, Body: %s", resp.StatusCode, string(body))
		}
	})
}
