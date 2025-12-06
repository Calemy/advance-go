package main

import (
	"io"
	"log"
	"os"
)

var cursor string

func initCursor() {
	file, err := os.OpenFile("cursor.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	csr, err := io.ReadAll(file)
	if err != nil {
		log.Println("Encounterted error while reading cursor file")
		return
	}

	cursor = string(csr)
}
