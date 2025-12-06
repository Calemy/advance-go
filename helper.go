package main

import (
	"strconv"
	"strings"
)

func JoinInts(nums []int, sep string) string {
	var b strings.Builder

	for i, n := range nums {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(strconv.Itoa(n))
	}

	return b.String()
}

func Find[T any](items []T, predicate func(T) bool) (T, bool) {
	var zero T
	for _, item := range items {
		if predicate(item) {
			return item, true
		}
	}
	return zero, false
}
