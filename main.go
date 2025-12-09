package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
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

	StartWebhookWorker(scoreWebhook)
	userBatcher.Start(time.Second * 10)  // 6 * 50 Ratelimit -> 300
	trackBatcher.Start(time.Second * 15) // 4 * 50 Ratelimit -> 200 -> 500

	loadUsers()

	defer DB.Close()

	go fetchScores() // 4 * 1 Ratelimit -> 4 -> 504
	go updateEmptyUsers()
	go updateUsers()

	wg.Add(2)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			fetchScores()
			updateEmptyUsers()
		}
	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			updateUsers()

			embed := discordwebhook.Embed{
				Title:       "Scores collcted",
				Description: fmt.Sprintf("%d scores collected, %d users added", scoreCount, usersEdited),
				Color:       0x00ff00,
				Timestamp:   time.Now(),
			}

			hook := discordwebhook.Hook{
				Username:   "Advance",
				Avatar_url: "https://a.ppy.sh/9527931",
				Embeds:     []discordwebhook.Embed{embed},
			}

			go func(hook discordwebhook.Hook) {
				webhookQueue <- hook
			}(hook)

			scoreCount = 0
			usersEdited = 0
		}
	}()
	wg.Wait()
}
