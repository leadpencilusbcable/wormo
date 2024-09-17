package main

import (
	"log"
	"sync"
	"wormo/http"
	"wormo/websocket"
)

const ROWS int = 40
const COLS int = 30

func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		error := http.StartServer(8000, ROWS, COLS)

		if error != nil {
			log.Panic(error)
		}
	}()

	go func() {
		defer waitGroup.Done()

		websocket.StartServer(8001, ROWS, COLS)
	}()

	waitGroup.Wait()
}
