package main

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value V
	exp   time.Time
	slot  int
}

type TTLMap[K comparable, V any] struct {
	mu      sync.Mutex
	data    map[K]entry[V]
	buckets []map[K]struct{}
	ttl     time.Duration
	size    int64
}

func TimedCache[K comparable, V any](ttl time.Duration) *TTLMap[K, V] {
	t := &TTLMap[K, V]{
		data:    make(map[K]entry[V]),
		ttl:     ttl,
		buckets: make([]map[K]struct{}, int(ttl.Minutes())),
	}
	for i := range t.buckets {
		t.buckets[i] = make(map[K]struct{})
	}
	go t.expirer()
	return t
}

func (t *TTLMap[K, V]) Get(key K) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.data[key]
	if !ok {
		var zero V
		return zero, false
	}

	if time.Now().After(e.exp) {
		delete(t.data, key)
		var zero V
		return zero, false
	}

	return e.value, true
}

func (t *TTLMap[K, V]) Set(key K, value V, ttl time.Duration) {
	exp := time.Now().Add(ttl)
	slot := int(exp.Unix() / 60 % int64(t.ttl.Minutes()))

	t.mu.Lock()
	defer t.mu.Unlock()

	if old, ok := t.data[key]; ok {
		delete(t.buckets[old.slot], key)
	}

	t.data[key] = entry[V]{value: value, exp: exp, slot: slot}
	t.buckets[slot][key] = struct{}{}
}

func (t *TTLMap[K, V]) expirer() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		slot := int(now.Unix() / 60 % int64(t.ttl.Minutes()))

		t.mu.Lock()
		for k := range t.buckets[slot] {
			if e, ok := t.data[k]; ok && now.After(e.exp) {
				delete(t.data, k)
			}
		}
		clear(t.buckets[slot])
		t.mu.Unlock()
	}
}

var scoreCache = TimedCache[int, struct{}](time.Hour * 24)
