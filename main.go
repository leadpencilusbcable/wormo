package main

import (
	"log"
	"sync"
	"wormo/http"
)

const ROWS uint8 = 40
const COLS uint8 = 30

func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()

		server, error := http.NewServer(
			8000,
			ROWS,
			COLS,
			"./public/pages/game.html",
			"./public/pages/error.html",
			"./public/pages/pagenotfound.html",
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

		//websocket.StartServer(8001, ROWS, COLS)
	}()

	waitGroup.Wait()
}
