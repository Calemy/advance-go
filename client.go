package main

import (
	"bufio"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	http       *http.Client
	localLimit *rate.Limiter
	remoteRL   *RemoteRL
	maxLimit   int
	inflight   chan struct{}
}

func NewLimitedClient(rps int) *Client {
	initProxies()

	var transport *http.Transport

	if os.Getenv("ENABLE_PROXY") == "true" && len(proxy.proxies) != 0 {
		transport = &http.Transport{
			Proxy:             proxy.NextProxy,
			DisableKeepAlives: true,
		}
	}

	return &Client{
		http: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},

		localLimit: rate.NewLimiter(rate.Limit(rps), rps),
		remoteRL:   NewRemoteRL(),
		inflight:   make(chan struct{}, 20),
	}

}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.inflight <- struct{}{}
	defer func() { <-c.inflight }()
	c.remoteRL.Check()

	if err := c.localLimit.Wait(req.Context()); err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		f, _ := os.OpenFile("error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		w := bufio.NewWriter(f)
		w.WriteString(err.Error() + "\n")

		w.Flush()
		f.Close()
		return nil, err
	}

	c.UpdateLimit(resp)

	if resp.StatusCode == 429 {
		c.remoteRL.TriggerFixed(time.Hour)
		log.Printf("Received 429, waiting for 1 hour")
		return nil, errors.New("remote rate limit reached (429)")
	}

	return resp, nil
}

type RemoteRL struct {
	mu        sync.Mutex
	waitUntil time.Time
	waiting   bool
	cond      *sync.Cond
	timer     *time.Timer
	remaining int
}

func NewRemoteRL() *RemoteRL {
	rl := &RemoteRL{}
	rl.cond = sync.NewCond(&rl.mu)
	return rl
}

func (rl *RemoteRL) Check() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.waiting {
		return
	}

	rl.cond.Wait()
}

func (rl *RemoteRL) TriggerFixed(d time.Duration) {
	rl.mu.Lock()

	if rl.waiting {
		rl.mu.Unlock()
		return
	}

	// Cancel any previous timer
	if rl.timer != nil {
		rl.timer.Stop()
	}

	rl.waiting = true
	rl.waitUntil = time.Now().Add(d)
	rl.timer = time.NewTimer(d)

	// Dedicated wait goroutine
	go func() {
		<-rl.timer.C
		rl.mu.Lock()
		rl.waiting = false
		rl.mu.Unlock()
		rl.cond.Broadcast()
	}()

	rl.mu.Unlock()
}

func (c *Client) UpdateLimit(resp *http.Response) {
	limitStr := resp.Header.Get("X-RateLimit-Limit")
	remainStr := resp.Header.Get("X-RateLimit-Remaining")
	// resetStr := resp.Header.Get("X-Ratelimit-Reset")

	if limitStr == "" || remainStr == "" {
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return
	}

	remain, err := strconv.Atoi(remainStr)
	if err != nil {
		return
	}

	if c.maxLimit == 0 {
		c.maxLimit = limit
	}

	if remain > c.maxLimit {
		c.maxLimit = remain
	}

	rl := c.remoteRL
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if remain > rl.remaining {
		if remain > (c.maxLimit / 2) {
			rl.remaining = remain
			rl.waiting = false
			if rl.timer != nil {
				rl.timer.Stop()
				rl.cond.Broadcast()
			}
		}
		return
	}

	rl.remaining = remain

	if remain < 100 {
		resetDur := 60 * time.Second
		rl.waiting = true
		rl.waitUntil = time.Now().Add(resetDur)

		if rl.timer != nil {
			rl.timer.Stop()
		}

		rl.timer = time.NewTimer(resetDur)

		go func() {
			<-rl.timer.C
			rl.mu.Lock()
			rl.waiting = false
			rl.mu.Unlock()
			rl.cond.Broadcast()
		}()
	}
}
