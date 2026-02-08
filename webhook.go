package main

import (
	"encoding/json"
	"log"
	"time"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
)

type WebhookWorker struct {
	Link   string
	queue  chan discordwebhook.Hook
	Ticker *time.Ticker
}

func NewWebhookWorker(link string) *WebhookWorker {
	return &WebhookWorker{
		Link:   link,
		queue:  make(chan discordwebhook.Hook),
		Ticker: time.NewTicker(time.Second * 3),
	}
}

func (w *WebhookWorker) Start() {
	go func() {
		for payload := range w.queue {
			<-w.Ticker.C
			_ = SendEmbed(w.Link, payload)
		}
	}()
}

func (w *WebhookWorker) Queue(hook discordwebhook.Hook) {
	w.queue <- hook
}

func SendEmbed(link string, hook discordwebhook.Hook) error {
	if link == "" {
		return nil
	}
	payload, err := json.Marshal(hook)
	if err != nil {
		log.Fatal(err)
	}
	err = discordwebhook.ExecuteWebhook(link, payload)
	return err
}
