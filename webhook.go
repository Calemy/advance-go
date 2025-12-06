package main

import (
	"encoding/json"
	"log"
	"time"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
)

var webhookQueue = make(chan discordwebhook.Hook)
var ticker = time.NewTicker(time.Second * 3)

func StartWebhookWorker(link string) {
	if link == "" {
		return
	}
	go func() {
		for payload := range webhookQueue {
			<-ticker.C
			_ = SendEmbed(link, payload)
		}
	}()
}

func SendEmbed(link string, hook discordwebhook.Hook) error {
	payload, err := json.Marshal(hook)
	if err != nil {
		log.Fatal(err)
	}
	err = discordwebhook.ExecuteWebhook(link, payload)
	return err

}
