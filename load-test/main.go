package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	CSVRecord struct {
		ServiceID   int
		ServiceName string
		Intent      string
	}

	Response struct {
		Data  ResponseData `json:"data"`
		Error string       `json:"error"`
	}

	ResponseData struct {
		ServiceID   int    `json:"service_id"`
		ServiceName string `json:"service_name"`
	}

	Result struct {
		Success bool
		Record  CSVRecord
		Error   string
		Latency time.Duration
	}

	OutputReport struct {
		TotalRequests int     `json:"total_requests"`
		Timestamp     string  `json:"timestamp"`
		ElapsedTime   string  `json:"elapsed_time"`
		TotalSuccess  int     `json:"total_success"`
		TotalFailed   int     `json:"total_failed"`
		SuccessRate   float64 `json:"success_rate"`
		FailureRate   float64 `json:"failure_rate"`
		FastestTime   string  `json:"fastest_time,omitempty"`
		SlowestTime   string  `json:"slowest_time,omitempty"`
		AverageTime   string  `json:"average_time,omitempty"`
	}
)

const (
	clientTimeout = 20 * time.Second
	numWorkers    = 20
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <csv_file> <endpoint_url> <output_result>")
		os.Exit(1)
	}

	csvFile := os.Args[1]
	endpointURL := os.Args[2]
	outputFile := os.Args[3]

	records, err := readCSV(csvFile)
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d records\n", len(records))

	jobs := make(chan CSVRecord, len(records))
	results := make(chan Result, len(records))

	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: clientTimeout,
	}

	sw := &Stopwatch{}
	sw.Start()

	defer sw.Stop()

	for i := range numWorkers {
		wg.Go(func() {
			worker(i+1, client, endpointURL, jobs, results)
		})
	}

	go func() {
		for i, record := range records {
			fmt.Printf("Queuing record %d: %s\n", i+1, record.ServiceName)
			jobs <- record
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var successCount int
	var failureCount int
	var fastestTime, slowestTime time.Duration
	var totalLatency time.Duration

	for result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}

		totalLatency += result.Latency

		if fastestTime == 0 || result.Latency < fastestTime {
			fastestTime = result.Latency
		}

		if result.Latency > slowestTime {
			slowestTime = result.Latency
		}
	}

	total := successCount + failureCount
	successRate := float64(successCount) / float64(total) * 100
	failureRate := float64(failureCount) / float64(total) * 100

	report := OutputReport{
		TotalRequests: total,
		ElapsedTime:   sw.FormatElapsed(),
		Timestamp:     time.Now().Format(time.RFC3339),
		TotalSuccess:  successCount,
		TotalFailed:   failureCount,
		SuccessRate:   successRate,
		FailureRate:   failureRate,
		FastestTime:   fmt.Sprintf("%dms", fastestTime.Milliseconds()),
		SlowestTime:   fmt.Sprintf("%dms", slowestTime.Milliseconds()),
		AverageTime:   fmt.Sprintf("%dms", (totalLatency / time.Duration(total)).Milliseconds()),
	}

	err = saveReportToFile(report, outputFile)
	if err != nil {
		fmt.Printf("Error saving report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Results saved to %s\n", outputFile)
}

func readCSV(filename string) ([]CSVRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var records []CSVRecord
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) < 3 {
			continue
		}

		// skip header
		if record[0] == "service_id" {
			continue
		}

		serviceID, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			panic(err)
		}

		records = append(records, CSVRecord{
			ServiceID:   serviceID,
			ServiceName: strings.TrimSpace(record[1]),
			Intent:      strings.TrimSpace(record[2]),
		})
	}

	return records, nil
}

func saveReportToFile(report OutputReport, filename string) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func worker(id int, client *http.Client, endpointURL string, jobs <-chan CSVRecord, results chan<- Result) {
	for record := range jobs {
		fmt.Printf("Worker %d processing: %s\n", id, record.ServiceName)

		startTime := time.Now()

		success := processRecord(client, endpointURL, record)
		elapsedTime := time.Since(startTime)

		results <- Result{
			Success: success,
			Record:  record,
			Latency: elapsedTime,
		}
	}
}

func processRecord(client *http.Client, endpointURL string, record CSVRecord) bool {
	payload := map[string]string{
		"intent": record.Intent,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return false
	}

	resp, err := client.Post(endpointURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return false
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API error: %s\n", response.Error)
		return false
	}

	if response.Data.ServiceID != record.ServiceID || response.Data.ServiceName != record.ServiceName {
		fmt.Printf("Validation failed for intent %q - Expected: ID=%d, Name=%s | Got: ID=%d, Name=%s\n",
			record.Intent, record.ServiceID, record.ServiceName, response.Data.ServiceID, response.Data.ServiceName)
		return false
	}

	fmt.Printf("Success - ID=%d, Name=%s\n", response.Data.ServiceID, response.Data.ServiceName)
	return true
}

// Stopwatch struct
type Stopwatch struct {
	start    time.Time
	duration time.Duration
	running  bool
}

// Start the stopwatch
func (sw *Stopwatch) Start() {
	if !sw.running {
		sw.start = time.Now()
		sw.running = true
	}
}

// Stop the stopwatch
func (sw *Stopwatch) Stop() {
	if sw.running {
		sw.duration += time.Since(sw.start)
		sw.running = false
	}
}

// Reset the stopwatch
func (sw *Stopwatch) Reset() {
	sw.start = time.Time{}
	sw.duration = 0
	sw.running = false
}

// Elapsed time of the stopwatch
func (sw *Stopwatch) Elapsed() time.Duration {
	if sw.running {
		return sw.duration + time.Since(sw.start)
	}
	return sw.duration
}

// FormatElapsed returns the elapsed time in mm:ss format
func (sw *Stopwatch) FormatElapsed() string {
	elapsed := sw.Elapsed()
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
