package main

import (
	"net/http"
	"os"
	"sync"
	"time"

	_ "net/http/pprof"

	_ "github.com/joho/godotenv/autoload"
)

var scoreWebhook string

func main() {
	var wg sync.WaitGroup

	InitDB(os.Getenv("POSTGRES_URL"))
	defer DB.Close()

	initCursor()
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	if os.Getenv("ENABLE_WEBHOOK") == "true" {
		scoreWebhook = os.Getenv("SCORES_WEBHOOK")
	}
	StartWebhookWorker(scoreWebhook)

	userUpdater.Workers(20)
	userUpdater.Start()

	loadUsers()

	fetchScores() // 4 * 1 Ratelimit -> 4 -> 604

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			fetchScores()
		}
	}()

	select {} // block forever (or start server)
}
