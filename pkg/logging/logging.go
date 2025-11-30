package pkg

import (
	"log"
	"os"
)

// A buffered channel to hold logs waiting to be written
var LogChannel = make(chan string, 100)

// Call this ONCE in your main.go
func StartLoggerWorker() {
	go func() {
		f, err := os.OpenFile("log.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		// This loop runs forever, processing one log at a time
		for msg := range LogChannel {
			if _, err := f.WriteString(msg + "\n"); err != nil {
				log.Println("Failed to write log:", err)
			}
		}
	}()
}
