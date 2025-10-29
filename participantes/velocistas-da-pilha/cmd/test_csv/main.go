package main
import (
    "fmt"
    "log"
    "velocistas_da_pilha/internal/storage"
)
func main() {
    intents, err := storage.LoadIntentsCSV("assets/intents_pre_loaded.csv")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Carregadas %d intenções.\n", len(intents))
    fmt.Printf("📄 Primeira: %+v\n", intents[0])
}