package websocket

import (
	"io"
	"log"
	"strconv"
	"strings"

	"golang.org/x/net/websocket"
)

const (
	eventInit            = "INIT"
	eventNewWorm         = "NEW"
	eventMove            = "MOVE"
	eventConsumeFood     = "CONSUMEFOOD"
	eventSpawnFood       = "SPAWNFOOD"
	eventSpawnBomb       = "SPAWNBOMB"
	eventDetonateBomb    = "DETBOMB"
	eventExtend          = "EXTEND"
	eventChangeDirection = "CHANGEDIR"
	eventDisconnect      = "DISCONNECT"
)

func (server *Server) handleChangeDir(initiatorId string, dir string) {
	server.mu.Lock()
	server.worms[initiatorId].direction = dir
	server.mu.Unlock()
}

func (server *Server) handleInit(initiator *websocket.Conn, initiatorId string) {
	msg := eventInit + "\n"

	newWormMsg := initiatorId + "," + positionsToString(server.worms[initiatorId].positions)
	msg += newWormMsg

	existingWormsMsg := ""

	for id, worm := range server.worms {
		if id != initiatorId {
			existingWormsMsg += id + "," + positionsToString(worm.positions) + "\n"
		}
	}

	if existingWormsMsg != "" {
		msg += "|" + existingWormsMsg[:len(existingWormsMsg)-1]
	} else {
		msg += "|"
	}

	foodPositionsMsg := ""

	//find faster way of doing this
	for x := 0; x < server.gridWidth; x++ {
		for y := 0; y < server.gridHeight; y++ {
			if server.grid[x][y].food {
				foodPositionsMsg += positionToString(&pos{x, y}) + ","
			}
		}
	}

	if foodPositionsMsg != "" {
		msg += "|" + foodPositionsMsg[:len(foodPositionsMsg)-1]
	} else {
		msg += "|"
	}

	bombPositionsMsg := ""

	server.mu.RLock()

	for id, bomb := range server.bombs {
		bombPositionsMsg += id + "," + strconv.FormatInt(int64(bomb.timeToDetonation), 10) + "," + positionToString(&bomb.bombPosition) + "," + positionsToString(bomb.positions) + "\n"
	}

	server.mu.RUnlock()

	if bombPositionsMsg != "" {
		msg += "|" + bombPositionsMsg[:len(bombPositionsMsg)-1]
	} else {
		msg += "|"
	}

	initiator.Write([]byte(msg))
	server.broadcastExcept([]byte(eventNewWorm+"\n"+newWormMsg), initiator)
}

func (server *Server) removePlayer(ws *websocket.Conn) {
	server.mu.RLock()

	id := server.wormConns[ws]
	positions := server.worms[id].positions

	server.mu.RUnlock()

	server.mu.Lock()

	for _, pos := range positions {
		server.grid[pos.x][pos.y].worm = ""
	}

	delete(server.worms, id)
	delete(server.wormConns, ws)

	server.mu.Unlock()

	server.broadcast([]byte(eventDisconnect + "\n" + id))
}

func (server *Server) readFromConnection(ws *websocket.Conn) {
	server.mu.RLock()
	id := server.wormConns[ws]
	server.mu.RUnlock()

	buffer := make([]byte, 1024)

	for {
		length, error := ws.Read(buffer)

		if error != nil {
			if error == io.EOF {
				server.removePlayer(ws)
				break
			}

			log.Println("Read error: ", error)
			continue
		}

		msg := string(buffer[:length])
		log.Println("msg from client: ", msg)

		eventDataSplitIndex := strings.Index(msg, "\n")

		//the name of the event, eg. INIT, NEW
		var event string
		//the data of the event, eg. worm positions
		var data string

		if eventDataSplitIndex != -1 {
			event = msg[0:eventDataSplitIndex]
			data = msg[eventDataSplitIndex+1:]
		} else {
			event = msg
		}

		switch event {
		case eventInit:
			{
				server.handleInit(ws, id)
			}
		case eventChangeDirection:
			{
				server.handleChangeDir(id, data)
			}
		}
	}
}

func (server *Server) handle(ws *websocket.Conn) {
	log.Println("Incoming connection: ", ws.RemoteAddr())

	server.mu.Lock()

	id := server.newWorm()
	server.wormConns[ws] = id

	server.mu.Unlock()

	server.readFromConnection(ws)
}
