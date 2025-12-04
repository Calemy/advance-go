package main

import (
	"errors"
)

var ErrFetch = errors.New("something went wrong while fetching")
var ErrNotFound = errors.New("this content could not be found")

func Error(message string) map[string]string {
	return map[string]string{
		"error": message,
	}
}
