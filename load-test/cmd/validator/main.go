package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TestResult represents the structure of the JSON test results
type TestResult struct {
	TotalRequests int     `json:"total_requests"`
	Timestamp     string  `json:"timestamp"`
	ElapsedTime   string  `json:"elapsed_time"`
	TotalSuccess  int     `json:"total_success"`
	TotalFailed   int     `json:"total_failed"`
	SuccessRate   float64 `json:"success_rate"`
	FailureRate   float64 `json:"failure_rate"`
	FastestTime   string  `json:"fastest_time"`
	SlowestTime   string  `json:"slowest_time"`
	AverageTime   string  `json:"average_time"`
}

// ParticipantResult holds combined results for a participant
type ParticipantResult struct {
	Name         string
	Test93       *TestResult
	Test80       *TestResult
	TotalSuccess int
	TotalFailed  int
	AvgTime93    float64 // in milliseconds
	AvgTime80    float64 // in milliseconds
	Score        float64
}

func main() {
	participantesPath := flag.String("path", "../../participantes", "Path to participantes folder")
	outputPath := flag.String("output", "results.html", "Output HTML file path")
	flag.Parse()

	if _, err := os.Stat(*participantesPath); os.IsNotExist(err) {
		fmt.Printf("Error: Path '%s' does not exist\n", *participantesPath)
		os.Exit(1)
	}

	participants, err := readAllParticipants(*participantesPath)
	if err != nil {
		fmt.Printf("Error reading participants: %v\n", err)
		os.Exit(1)
	}

	if len(participants) == 0 {
		fmt.Println("No participants with valid results found")
		os.Exit(0)
	}

	// Calculate scores and rank
	for i := range participants {
		participants[i].Score = calculateScore(&participants[i])
	}

	// Sort by score (higher is better)
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].Score > participants[j].Score
	})

	// Generate HTML report
	err = generateHTMLReport(participants, *outputPath)
	if err != nil {
		fmt.Printf("Error generating HTML report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Report generated successfully: %s\n", *outputPath)
}

func readAllParticipants(basePath string) ([]ParticipantResult, error) {
	var participants []ParticipantResult

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		participantName := entry.Name()
		resultsPath := filepath.Join(basePath, participantName, "results")

		// Check if results directory exists
		if _, err := os.Stat(resultsPath); os.IsNotExist(err) {
			fmt.Printf("Warning: No results folder for participant '%s', skipping\n", participantName)
			continue
		}

		// Read test results
		test93, err93 := readTestResult(filepath.Join(resultsPath, "93.json"))
		test80, err80 := readTestResult(filepath.Join(resultsPath, "80.json"))

		// Skip if both files are missing
		if err93 != nil && err80 != nil {
			fmt.Printf("Warning: No valid test results for participant '%s', skipping\n", participantName)
			continue
		}

		participant := ParticipantResult{
			Name:   participantName,
			Test93: test93,
			Test80: test80,
		}

		// Calculate aggregated metrics
		if test93 != nil {
			participant.TotalSuccess += test93.TotalSuccess
			participant.TotalFailed += test93.TotalFailed
			participant.AvgTime93 = parseTimeMs(test93.AverageTime)
		}

		if test80 != nil {
			participant.TotalSuccess += test80.TotalSuccess
			participant.TotalFailed += test80.TotalFailed
			participant.AvgTime80 = parseTimeMs(test80.AverageTime)
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func readTestResult(filePath string) (*TestResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result TestResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// parseTimeMs extracts milliseconds from time strings like "3938ms"
func parseTimeMs(timeStr string) float64 {
	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.TrimSuffix(timeStr, "ms")

	val, err := strconv.ParseFloat(timeStr, 64)
	if err != nil {
		return 0
	}

	return val
}

// calculateScore computes a ranking score based on success rate, failures, and response time
func calculateScore(p *ParticipantResult) float64 {
	// Scoring weights (adjusted for realistic values)
	const (
		successWeight = 10.0 // Each success worth 10 points
		failureWeight = 50.0 // Each failure costs 50 points
		timeWeight    = 0.01 // Time penalty: avg_time (in ms) / 100
	)

	score := 0.0

	// Higher success count is better
	score += float64(p.TotalSuccess) * successWeight

	// Lower failure count is better (subtract from score)
	score -= float64(p.TotalFailed) * failureWeight

	// Lower average time is better
	// Calculate average of both tests (in milliseconds)
	avgTime := (p.AvgTime93 + p.AvgTime80) / 2.0
	if avgTime > 0 {
		// Time penalty scaled down to not dominate the score
		score -= avgTime * timeWeight
	}

	return score
}

func generateHTMLReport(participants []ParticipantResult, outputPath string) error {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatTime": formatTime,
	}

	tmpl := template.Must(template.New("report").Funcs(funcMap).Parse(htmlTemplate))

	data := struct {
		Participants []ParticipantResult
		GeneratedAt  string
	}{
		Participants: participants,
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Load Test Results - Participant Rankings</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            min-height: 100vh;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }

        .header .subtitle {
            font-size: 1.1em;
            opacity: 0.9;
        }

        .generated-at {
            font-size: 0.9em;
            opacity: 0.8;
            margin-top: 10px;
        }

        .scoring-info {
            background: #f8f9fa;
            padding: 30px;
            border-bottom: 3px solid #e9ecef;
        }

        .scoring-info h2 {
            color: #495057;
            margin-bottom: 15px;
            font-size: 1.5em;
        }

        .scoring-formula {
            background: white;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            margin: 15px 0;
        }

        .scoring-formula code {
            background: #f1f3f5;
            padding: 2px 8px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            color: #495057;
        }

        .criteria-list {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
            margin-top: 15px;
        }

        .criteria-item {
            background: white;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #28a745;
        }

        .criteria-item.penalty {
            border-left-color: #dc3545;
        }

        .criteria-item.time {
            border-left-color: #ffc107;
        }

        .rankings {
            padding: 40px;
        }

        .rankings h2 {
            color: #495057;
            margin-bottom: 25px;
            font-size: 1.8em;
        }

        .ranking-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 40px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            border-radius: 8px;
            overflow: hidden;
        }

        .ranking-table thead {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .ranking-table th {
            padding: 15px;
            text-align: left;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85em;
            letter-spacing: 0.5px;
        }

        .ranking-table td {
            padding: 15px;
            border-bottom: 1px solid #e9ecef;
        }

        .ranking-table tbody tr:hover {
            background: #f8f9fa;
            transition: background 0.2s;
        }

        .rank-badge {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 40px;
            height: 40px;
            border-radius: 50%;
            font-weight: bold;
            font-size: 1.1em;
        }

        .rank-1 {
            background: linear-gradient(135deg, #FFD700, #FFA500);
            color: white;
            box-shadow: 0 4px 10px rgba(255, 215, 0, 0.4);
        }

        .rank-2 {
            background: linear-gradient(135deg, #C0C0C0, #A8A8A8);
            color: white;
            box-shadow: 0 4px 10px rgba(192, 192, 192, 0.4);
        }

        .rank-3 {
            background: linear-gradient(135deg, #CD7F32, #8B4513);
            color: white;
            box-shadow: 0 4px 10px rgba(205, 127, 50, 0.4);
        }

        .rank-other {
            background: #e9ecef;
            color: #495057;
        }

        .participant-name {
            font-weight: 600;
            color: #495057;
            font-size: 1.1em;
        }

        .metric {
            font-weight: 500;
        }

        .metric.success {
            color: #28a745;
        }

        .metric.failed {
            color: #dc3545;
        }

        .score {
            font-weight: bold;
            font-size: 1.2em;
            color: #667eea;
        }

        .details {
            margin-top: 40px;
        }

        .participant-card {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 25px;
            border: 2px solid #e9ecef;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .participant-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 20px rgba(0,0,0,0.1);
        }

        .participant-card-header {
            display: flex;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 2px solid #dee2e6;
        }

        .participant-card-header .rank-badge {
            margin-right: 15px;
        }

        .participant-card-header h3 {
            color: #495057;
            font-size: 1.5em;
        }

        .test-results {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
        }

        .test-result-box {
            background: white;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }

        .test-result-box h4 {
            color: #667eea;
            margin-bottom: 15px;
            font-size: 1.2em;
        }

        .test-result-box.no-data {
            border-left-color: #6c757d;
            opacity: 0.6;
        }

        .test-stat {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #f1f3f5;
        }

        .test-stat:last-child {
            border-bottom: none;
        }

        .test-stat-label {
            color: #6c757d;
            font-weight: 500;
        }

        .test-stat-value {
            color: #495057;
            font-weight: 600;
        }

        .combined-score {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 15px 20px;
            border-radius: 8px;
            margin-top: 20px;
            text-align: center;
            font-size: 1.3em;
            font-weight: bold;
        }

        @media (max-width: 768px) {
            .header h1 {
                font-size: 1.8em;
            }

            .criteria-list {
                grid-template-columns: 1fr;
            }

            .test-results {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏆 Load Test Results</h1>
            <div class="subtitle">Participant Rankings - Credsystem Hackathon 2025</div>
            <div class="generated-at">Generated at: {{.GeneratedAt}}</div>
        </div>

        <div class="scoring-info">
            <h2>📊 Scoring Criteria</h2>
            <div class="scoring-formula">
                <strong>Formula:</strong> <code>Score = (Total_Success × 10.0) - (Total_Failed × 50.0) - (Avg_Time_ms × 0.01)</code>
            </div>
            <div class="criteria-list">
                <div class="criteria-item">
                    <strong>✅ Success Weight:</strong> +10.0 points per success
                </div>
                <div class="criteria-item penalty">
                    <strong>❌ Failure Penalty:</strong> -50.0 points per failure
                </div>
                <div class="criteria-item time">
                    <strong>⏱️ Time Penalty:</strong> -0.01 points per millisecond
                </div>
            </div>
        </div>

        <div class="rankings">
            <h2>🏅 Rankings</h2>
            <table class="ranking-table">
                <thead>
                    <tr>
                        <th>Rank</th>
                        <th>Participant</th>
                        <th>Total Success</th>
                        <th>Total Failed</th>
                        <th>Avg Time (93)</th>
                        <th>Avg Time (80)</th>
                        <th>Final Score</th>
                    </tr>
                </thead>
                <tbody>
                    {{range $index, $p := .Participants}}
                    <tr>
                        <td>
                            {{if eq $index 0}}
                                <span class="rank-badge rank-1">{{add $index 1}}</span>
                            {{else if eq $index 1}}
                                <span class="rank-badge rank-2">{{add $index 1}}</span>
                            {{else if eq $index 2}}
                                <span class="rank-badge rank-3">{{add $index 1}}</span>
                            {{else}}
                                <span class="rank-badge rank-other">{{add $index 1}}</span>
                            {{end}}
                        </td>
                        <td><span class="participant-name">{{$p.Name}}</span></td>
                        <td><span class="metric success">{{$p.TotalSuccess}}</span></td>
                        <td><span class="metric failed">{{$p.TotalFailed}}</span></td>
                        <td>{{formatTime $p.AvgTime93}}</td>
                        <td>{{formatTime $p.AvgTime80}}</td>
                        <td><span class="score">{{printf "%.2f" $p.Score}}</span></td>
                    </tr>
                    {{end}}
                </tbody>
            </table>

            <h2>📋 Detailed Breakdown</h2>
            <div class="details">
                {{range $index, $p := .Participants}}
                <div class="participant-card">
                    <div class="participant-card-header">
                        {{if eq $index 0}}
                            <span class="rank-badge rank-1">{{add $index 1}}</span>
                        {{else if eq $index 1}}
                            <span class="rank-badge rank-2">{{add $index 1}}</span>
                        {{else if eq $index 2}}
                            <span class="rank-badge rank-3">{{add $index 1}}</span>
                        {{else}}
                            <span class="rank-badge rank-other">{{add $index 1}}</span>
                        {{end}}
                        <h3>{{$p.Name}}</h3>
                    </div>

                    <div class="test-results">
                        {{if $p.Test93}}
                        <div class="test-result-box">
                            <h4>Test 93 Results</h4>
                            <div class="test-stat">
                                <span class="test-stat-label">Total Requests:</span>
                                <span class="test-stat-value">{{$p.Test93.TotalRequests}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Success:</span>
                                <span class="test-stat-value" style="color: #28a745;">{{$p.Test93.TotalSuccess}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Failed:</span>
                                <span class="test-stat-value" style="color: #dc3545;">{{$p.Test93.TotalFailed}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Success Rate:</span>
                                <span class="test-stat-value">{{printf "%.1f" $p.Test93.SuccessRate}}%</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Average Time:</span>
                                <span class="test-stat-value">{{$p.Test93.AverageTime}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Fastest Time:</span>
                                <span class="test-stat-value">{{$p.Test93.FastestTime}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Slowest Time:</span>
                                <span class="test-stat-value">{{$p.Test93.SlowestTime}}</span>
                            </div>
                        </div>
                        {{else}}
                        <div class="test-result-box no-data">
                            <h4>Test 93 Results</h4>
                            <p>No data available</p>
                        </div>
                        {{end}}

                        {{if $p.Test80}}
                        <div class="test-result-box">
                            <h4>Test 80 Results</h4>
                            <div class="test-stat">
                                <span class="test-stat-label">Total Requests:</span>
                                <span class="test-stat-value">{{$p.Test80.TotalRequests}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Success:</span>
                                <span class="test-stat-value" style="color: #28a745;">{{$p.Test80.TotalSuccess}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Failed:</span>
                                <span class="test-stat-value" style="color: #dc3545;">{{$p.Test80.TotalFailed}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Success Rate:</span>
                                <span class="test-stat-value">{{printf "%.1f" $p.Test80.SuccessRate}}%</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Average Time:</span>
                                <span class="test-stat-value">{{$p.Test80.AverageTime}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Fastest Time:</span>
                                <span class="test-stat-value">{{$p.Test80.FastestTime}}</span>
                            </div>
                            <div class="test-stat">
                                <span class="test-stat-label">Slowest Time:</span>
                                <span class="test-stat-value">{{$p.Test80.SlowestTime}}</span>
                            </div>
                        </div>
                        {{else}}
                        <div class="test-result-box no-data">
                            <h4>Test 80 Results</h4>
                            <p>No data available</p>
                        </div>
                        {{end}}
                    </div>

                    <div class="combined-score">
                        Combined Score: {{printf "%.2f" $p.Score}} points
                    </div>
                </div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>`

func formatTime(ms float64) string {
	if ms == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.0fms", ms)
}
