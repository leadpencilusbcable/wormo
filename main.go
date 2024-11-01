package main

import (
	"flag"
	"log"
	"sync"
	"wormo/http"
	"wormo/websocket"
)

const ROWS uint8 = 40
const COLS uint8 = 30
const LEVEL_MULTIPLIER uint8 = 1

func main() {
	httpPort := flag.Int("http-port", 8000, "port number for http connections")
	wsPort := flag.Int("ws-port", 8001, "port number for ws connections")

	flag.Parse()

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()

		server, error := http.NewServer(
			uint16(*httpPort),
			uint16(*wsPort),
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

		server := websocket.NewServer(uint16(*wsPort), ROWS, COLS, LEVEL_MULTIPLIER)

		server.Server.ListenAndServe()
	}()

	waitGroup.Wait()
}
