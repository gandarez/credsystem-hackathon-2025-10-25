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

type ServiceData struct {
	// Add any fields your API returns if you want to inspect them
	// Example:
	// Intent string `json:"intent"`
}

type Message struct {
	Success bool        `json:"success"`
	Data    ServiceData `json:"data"`
	Error   string      `json:"error"`
}

var (
	mu           sync.Mutex
	totalTime    time.Duration
	totalCount   int
	successCount int
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

	blocks := bytes.Split(data, []byte("\n\n"))
	var payloads []string
	for _, b := range blocks {
		s := strings.TrimSpace(string(b))
		if s != "" {
			payloads = append(payloads, s)
		}
	}

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

func saveResponse(intent string, response string, duration time.Duration, success bool, errMsg string) {
	mu.Lock()
	defer mu.Unlock()

	f, err := os.OpenFile("responses.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("❌ Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	status := "❌ FAILED"
	if success {
		status = "✅ SUCCESS"
		successCount++
	}

	entry := fmt.Sprintf(
		"---\nRequest: %s\nResponse: %s\nResult: %s\nError: %s\nTime: %v\n\n",
		strings.TrimSpace(intent),
		strings.TrimSpace(response),
		status,
		errMsg,
		duration,
	)
	if _, err := f.WriteString(entry); err != nil {
		fmt.Printf("❌ Error writing to file: %v\n", err)
	}

	totalTime += duration
	totalCount++
}

func worker(wg *sync.WaitGroup, client *http.Client, url string, jobs <-chan string, id int) {
	defer wg.Done()
	for body := range jobs {
		start := time.Now()

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
		if err != nil {
			fmt.Printf("[w%d] request create error: %v\n", id, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		duration := time.Since(start)
		if err != nil {
			fmt.Printf("[w%d] post error (%.2fs): %v\n", id, duration.Seconds(), err)
			saveResponse(body, "", duration, false, err.Error())
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var msg Message
		success := false
		errMsg := ""
		if err := json.Unmarshal(respBody, &msg); err == nil {
			success = msg.Success
			errMsg = msg.Error
		} else {
			errMsg = "invalid JSON response"
		}

		fmt.Printf("[w%d] %s -> success=%v (%.2fs)\n", id, url, success, duration.Seconds())
		saveResponse(body, string(respBody), duration, success, errMsg)
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

	os.Remove("responses.txt")
	startAll := time.Now()

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

	elapsed := time.Since(startAll)

	avg := time.Duration(0)
	if totalCount > 0 {
		avg = totalTime / time.Duration(totalCount)
	}

	fmt.Println("✅ Done.")
	fmt.Printf("🕒 Total time: %v\n", elapsed)
	fmt.Printf("📈 Average response time: %v\n", avg)
	fmt.Printf("📊 Success: %d / %d (%.2f%%)\n", successCount, totalCount, float64(successCount)/float64(totalCount)*100)
	fmt.Printf("📦 Responses saved to responses.txt\n")
}
