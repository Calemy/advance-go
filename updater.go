package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Updater struct {
	mu   sync.Mutex
	list map[int]struct{}

	priority chan int   // used for empty users and retries
	in       chan int   // only take in users in a queue
	out      chan []int // only give out batches
}

var updater = &Updater{
	list:     make(map[int]struct{}),
	priority: make(chan int),
	in:       make(chan int),
	out:      make(chan []int),
}

func (u *Updater) Queue(id int, priority bool) {
	u.mu.Lock()
	if _, ok := u.list[id]; ok {
		u.mu.Unlock()
		return
	}
	u.list[id] = struct{}{}
	u.mu.Unlock()

	if priority {
		u.priority <- id
	} else {
		u.in <- id
	}
}

func (u *Updater) Remove(id int) {
	u.mu.Lock()
	delete(u.list, id)
	u.mu.Unlock()
}

func (u *Updater) Start(timeout, cooldown time.Duration) {
	go func() {
		limiter := time.NewTicker(cooldown)
		defer limiter.Stop()

		batch := make([]int, 0, 50)
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		flush := func() {
			if len(batch) == 0 {
				return
			}

			<-limiter.C

			out := make([]int, len(batch))
			copy(out, batch)

			u.out <- out
			batch = batch[:0]
		}

		for {
			if len(batch) == 50 {
				flush()
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
				continue
			}

			select {
			case id := <-u.priority:
				batch = append(batch, id)
				continue
			default:
			}

			select {
			case id := <-u.in:
				batch = append(batch, id)

			case <-timer.C:
				flush()
				timer.Reset(timeout)
			}
		}
	}()
}

func (u *Updater) Workers(n int) {
	for i := 0; i < n; i++ {
		go u.worker(i)
	}
}

func (u *Updater) worker(id int) {
	for batch := range u.out {
		if err := UpdateUsers(batch); err != nil {
			log.Printf("Worker %d failed processing batch: %v", id, err)
			for _, user := range batch {
				u.Remove(user)
				go u.Queue(user, true)
			}
			continue
		}
	}
}

func UpdateUsers(users []int) error {
	body, err := Fetch(fmt.Sprintf("/users?include_variant_statistics=true&ids[]=%s", JoinInts(users, "&ids[]=")))
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return errors.New("Empty body")
	}

	var resp UsersResponse

	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	var wg sync.WaitGroup

	userMap := make(map[int]*UserExtended, len(resp.Users))

	// Populate data to check later if anyone has been restricted
	for i := range resp.Users {
		u := &resp.Users[i]
		userMap[u.ID] = u
	}

	for _, user := range users {
		u, ok := userMap[user]
		if !ok {
			u := &UserExtended{ID: user}
			u.Restrict()
			updater.Remove(user)
			continue
		}

		wg.Add(1)
		go func(u *UserExtended) {
			defer wg.Done()
			u.Update()
			for mode, stats := range *u.StatisticsRulesets {
				if !stats.IsRanked || stats.PP == 0 {
					continue
				}
				stats.UpdateHistory(u.ID, ModeInt(mode))
			}
			// log.Printf("Finished updating %s (%d)", u.Username, u.ID)
			updater.Remove(user)
		}(u)
	}

	wg.Wait()

	return nil
}
