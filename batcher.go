package main

import (
	"sync"
	"time"
)

type Batcher[T any] struct {
	input     chan T
	batchSize int
	timeout   time.Duration
	flushFn   func([]T)
}

func NewBatcher[T any](batchSize int, timeout time.Duration, flushFn func([]T)) *Batcher[T] {
	return &Batcher[T]{
		input:     make(chan T, 2000),
		batchSize: batchSize,
		timeout:   timeout,
		flushFn:   flushFn,
	}
}

func (b *Batcher[T]) Start(cooldown time.Duration) {
	queue := make([]T, 0)
	queueMu := sync.Mutex{}
	queueCond := sync.NewCond(&queueMu)

	go func() {
		for item := range b.input {
			queueMu.Lock()
			queue = append(queue, item)
			queueCond.Signal()
			queueMu.Unlock()
		}
	}()

	go func() {
		batch := make([]T, 0, b.batchSize)
		timer := time.NewTimer(b.timeout)
		lastFlush := time.Time{}

		flush := func() {
			now := time.Now()
			if !lastFlush.IsZero() {
				if since := now.Sub(lastFlush); since < cooldown {
					time.Sleep(cooldown - since)
				}
			}
			b.flushFn(batch)
			lastFlush = time.Now()
		}

		for {
			queueMu.Lock()

			for len(queue) == 0 {
				queueCond.Wait()
			}

			n := b.batchSize
			if len(queue) < n {
				n = len(queue)
			}
			batch = append(batch[:0], queue[:n]...)
			queue = queue[n:]

			queueMu.Unlock()

			flush()

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(b.timeout)
		}
	}()
}

func (b *Batcher[T]) Add(item T) {
	b.input <- item
}
