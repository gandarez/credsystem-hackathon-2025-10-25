package configs

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("./participantes/campeoes-do-canal/.env"); err != nil {
		panic(err)
	}
}

func GetPort() string {
	return os.Getenv("PORT")
}
