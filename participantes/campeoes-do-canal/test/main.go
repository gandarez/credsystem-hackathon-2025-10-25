package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func readPayloads(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Split by blank lines
	blocks := bytes.Split(data, []byte("\n\n"))
	var payloads []string
	for _, b := range blocks {
		s := strings.TrimSpace(string(b))
		if s != "" {
			payloads = append(payloads, s)
		}
	}

	// Fallback to line-by-line if only one payload
	if len(payloads) < 2 {
		if _, err := f.Seek(0, 0); err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				payloads = append(payloads, line)
			}
		}
	}

	// Validate JSON
	valid := make([]string, 0, len(payloads))
	for _, p := range payloads {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(p), &js); err == nil {
			valid = append(valid, p)
		} else {
			fmt.Printf("⚠️ Invalid JSON skipped: %s\n", p)
		}
	}

	return valid, nil
}

func worker(wg *sync.WaitGroup, client *http.Client, url string, jobs <-chan string, id int) {
	defer wg.Done()
	for body := range jobs {
		req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
		if err != nil {
			fmt.Printf("[w%d] request create error: %v\n", id, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[w%d] post error: %v\n", id, err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Printf("[w%d] %s -> %d\n", id, url, resp.StatusCode)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go test/payload.txt http://localhost:18020 [concurrency]")
		return
	}

	payloadFile := os.Args[1]
	host := os.Args[2]
	concurrency := 5
	if len(os.Args) >= 4 {
		fmt.Sscan(os.Args[3], &concurrency)
	}

	payloads, err := readPayloads(payloadFile)
	if err != nil {
		fmt.Printf("❌ Error reading payloads: %v\n", err)
		return
	}
	if len(payloads) == 0 {
		fmt.Println("❌ No payloads found in file")
		return
	}

	url := strings.TrimRight(host, "/") + "/api/find-service"
	fmt.Printf("🚀 Sending %d payloads to %s (concurrency=%d)\n", len(payloads), url, concurrency)

	jobs := make(chan string, len(payloads))
	client := &http.Client{Timeout: 15 * time.Second}
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(&wg, client, url, jobs, i+1)
	}

	for _, p := range payloads {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
	fmt.Println("✅ Done")
}
