package main

import (
	"os"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var scoreWebhook string

func main() {
	var wg sync.WaitGroup
	InitDB(os.Getenv("POSTGRES_URL"))
	initCursor()
	if os.Getenv("ENABLE_WEBHOOK") == "true" {
		scoreWebhook = os.Getenv("SCORES_WEBHOOK")
	}

	defer DB.Close()
	defer cursorFile.Close()
	if userCount == 0 {

	}

	if usersEdited == 0 {

	}

	if scoreCount == 0 {

	}

	fetchScores()

	wg.Add(2)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			fetchScores()
		}
	}()
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			fetchUsers()
			scoreCount = 0
			usersEdited = 0
		}
	}()
	wg.Wait()
}
