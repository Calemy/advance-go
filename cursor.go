package main

import (
	"io"
	"log"
	"os"
)

var cursorFile *os.File
var cursor string

func initCursor() {
	file, err := os.OpenFile("cursor.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	cursorFile = file
	csr, err := io.ReadAll(file)
	if err != nil {
		log.Println("Encounterted error while reading cursor file")
		return
	}

	cursor = string(csr)
}
