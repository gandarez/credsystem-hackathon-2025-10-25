package util

import (
	"encoding/csv"
	"os"
	"strings"
)

func DetermineService(intent string) (uint8, string) {
	file, err := os.Open("data/services.csv")
	if err != nil {
		return 0, "error_opening_csv"
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, "error_reading_csv"
	}

	intentLower := strings.ToLower(intent)

	for _, row := range records {
		if len(row) < 3 {
			continue
		}
		if strings.Contains(intentLower, strings.ToLower(row[1])) {
			return parseUint8(row[0]), row[2]
		}
	}
	return 0, "unknown_service"
}

func parseUint8(s string) uint8 {
	var id uint8
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			id = id*10 + uint8(ch-'0')
		}
	}
	return id
}
