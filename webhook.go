package main

import (
	"encoding/json"
	"log"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
)

func SendEmbed(link string, hook discordwebhook.Hook) error {
	payload, err := json.Marshal(hook)
	if err != nil {
		log.Fatal(err)
	}
	err = discordwebhook.ExecuteWebhook(link, payload)
	return err

}
