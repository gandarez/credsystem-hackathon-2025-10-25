package storage
import (
    "encoding/csv"
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
)
// IntentEntry representa uma linha do CSV com ID, nome e texto da intenção.
type IntentEntry struct {
    ServiceID   int
    ServiceName string
    Intent      string
}
// LoadIntentsCSV lê o arquivo de intenções e retorna todas as entradas.
func LoadIntentsCSV(path string) ([]IntentEntry, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("falha ao abrir CSV: %w", err)
    }
    defer file.Close()
    reader := csv.NewReader(file)
    reader.Comma = ';' // separador usado no arquivo
    reader.LazyQuotes = true
    // Ignorar cabeçalho
    if _, err := reader.Read(); err != nil {
        return nil, fmt.Errorf("falha ao ler cabeçalho: %w", err)
    }
    var intents []IntentEntry
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("erro lendo linha CSV: %w", err)
        }
        id, _ := strconv.Atoi(strings.TrimSpace(record[0]))
        entry := IntentEntry{
            ServiceID:   id,
            ServiceName: strings.TrimSpace(record[1]),
            Intent:      strings.TrimSpace(record[2]),
        }
        intents = append(intents, entry)
    }
    return intents, nil
}