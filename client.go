package main

import (
	"errors"
	"net/http"
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
}

func NewLimitedClient(rps int) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		localLimit: rate.NewLimiter(rate.Limit(rps), rps),
		remoteRL:   NewRemoteRL(),
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.remoteRL.Check()

	if err := c.localLimit.Wait(req.Context()); err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	c.UpdateLimit(resp)

	if resp.StatusCode == 429 {
		c.remoteRL.TriggerFixed(time.Hour)
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

	// If another goroutine hit remote RL, we wait here
	for rl.waiting {
		rl.cond.Wait()
	}

	// Fallback: woke too early
	now := time.Now()
	if now.Before(rl.waitUntil) {
		sleep := rl.waitUntil.Sub(now)
		rl.mu.Unlock()
		time.Sleep(sleep)
		return
	}

	rl.mu.Unlock()
}

func (rl *RemoteRL) TriggerFixed(d time.Duration) {
	rl.mu.Lock()

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
