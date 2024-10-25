package main

import (
	"log"
	"sync"
	"wormo/http"
	"wormo/websocket"
)

const ROWS uint8 = 40
const COLS uint8 = 30
const LEVEL_MULTIPLIER uint8 = 1

func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()

		server, error := http.NewServer(
			8000,
			ROWS,
			COLS,
			LEVEL_MULTIPLIER,
			"./public/pages/game.html",
			"./public/pages/error.html",
			"./public/pages/pagenotfound.html",
			"public/images",
			"public/styles",
			"public/scripts",
		)

		if error != nil {
			log.Panic(error)
		}

		server.Server.ListenAndServe()
	}()

	go func() {
		defer waitGroup.Done()

		server := websocket.NewServer(8001, ROWS, COLS, LEVEL_MULTIPLIER)

		server.Server.ListenAndServe()
	}()

	waitGroup.Wait()
}
