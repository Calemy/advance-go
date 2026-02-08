package main

import "sync"

type Queue struct {
	in       chan int
	priority chan int
	out      chan int

	flush func(id int, modes uint8) error

	mu    sync.Mutex
	cache map[int]uint8
}

func (q *Queue) Start() {
	go func() {
		for {
			select {
			case id := <-q.priority:
				q.out <- id
			case id := <-q.in:
				q.out <- id
			}
		}
	}()
}

func (q *Queue) Queue(id int, mode uint8, priority bool) {
	bit := uint8(1 << mode)

	q.mu.Lock()
	prev := q.cache[id]
	q.cache[id] = prev | bit
	q.mu.Unlock()

	if prev == 0 {
		if priority {
			q.priority <- id
		} else {
			q.in <- id
		}
	}
}

func (q *Queue) Remove(id int) {
	q.mu.Lock()
	delete(q.cache, id)
	q.mu.Unlock()
}

func (q *Queue) Workers(n int) {
	for i := 0; i < n; i++ {
		go q.worker()
	}
}

func (q *Queue) worker() {
	for id := range q.out {
		q.mu.Lock()
		modes := q.cache[id]
		delete(q.cache, id)
		q.mu.Unlock()

		if modes == 0 {
			continue
		}

		if err := q.flush(id, modes); err != nil {
			q.mu.Lock()
			q.cache[id] |= modes
			q.mu.Unlock()

			go func() {
				q.priority <- id
			}()
		}
	}
}

func createQueue(flush func(id int, modes uint8) error, slots, pSlots int) *Queue {
	q := &Queue{
		in:       make(chan int, slots),
		out:      make(chan int, 20),
		priority: make(chan int, pSlots),
		flush:    flush,
		cache:    make(map[int]uint8),
	}

	return q
}

var userUpdater = createQueue(updateUser, 512, 256)