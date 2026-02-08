package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "net/http/pprof"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
	_ "github.com/joho/godotenv/autoload"
)

var statsWebhook string
var includeFailed = 0

func main() {
	var wg sync.WaitGroup

	rpsStr := os.Getenv("REQUESTS_PER_SECOND")
	rps := 5
	if rpsStr != "" {
		parsed, err := strconv.Atoi(rpsStr)
		if err == nil && parsed != 0 {
			rps = parsed
		}
	}

	if os.Getenv("INCLUDE_FAILED") == "true" {
		includeFailed = 1
	}

	client = NewLimitedClient(rps)

	InitDB(os.Getenv("POSTGRES_URL"))
	defer DB.Close()

	initCursor()
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	if os.Getenv("ENABLE_WEBHOOK") == "true" {
		statsWebhook = os.Getenv("STATS_WEBHOOK")
	}
	StartWebhookWorker(statsWebhook)

	userUpdater.Workers(20)
	userUpdater.Start()

	loadUsers()
	loadQueue()

	fetchScores() // 4 * 1 Ratelimit -> 4 -> 604

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			fetchScores()
		}
	}()

	go func() {
		now := time.Now()
		nextHour := now.Truncate(time.Hour).Add(time.Hour)
		time.Sleep(time.Until(nextHour))

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			embed := discordwebhook.Embed{
				Title:       "Update Stats",
				Color:       0x86DC3D,
				Timestamp:   time.Now(),
				Footer: discordwebhook.Footer{
					Text: fmt.Sprintf("Users tracked: %d", userCount),
				},
				Fields: []discordwebhook.Field{
					{
						Name:   "Scores Stored",
						Value:  fmt.Sprintf("%d", scoreCount),
						Inline: true,
					},
					{
						Name:   "Stats Updated",
						Value:  fmt.Sprintf("%d", statsCount),
						Inline: true,
					},
				},
			}
			hook := discordwebhook.Hook{
				Username:   "Advance",
				Avatar_url: "https://a.ppy.sh/9527931",
				Embeds:     []discordwebhook.Embed{embed},
			}
			webhookQueue <- hook
			scoreCount = 0
			statsCount = 0
			<-ticker.C
		}
	}()

	select {} // block forever (or start server)
}
