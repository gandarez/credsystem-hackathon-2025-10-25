package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	// Try to load .env if present; don't panic when missing so the app can run
	// in environments where environment variables are provided externally.
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: .env not loaded: %v", err)
	}
}

func GetPort() string {
	return os.Getenv("PORT")
}
