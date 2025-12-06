package main

import (
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

func (b *Batcher[T]) Start() {
	go func() {
		batch := make([]T, 0, b.batchSize)
		timer := time.NewTimer(b.timeout)

		for {
			select {
			case item := <-b.input:
				batch = append(batch, item)

				if len(batch) >= b.batchSize {
					b.flushFn(batch)
					batch = batch[:0]
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(b.timeout)
				}

			case <-timer.C:
				if len(batch) > 0 {
					b.flushFn(batch)
					batch = batch[:0]
				}
				timer.Reset(b.timeout)
			}
		}
	}()
}

func (b *Batcher[T]) Add(item T) {
	b.input <- item
}
