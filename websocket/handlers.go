package websocket

import (
	"io"
	"log"
	"strings"

	"golang.org/x/net/websocket"
)

const (
	eventInit        = "INIT"
	eventNewWorm     = "NEW"
	eventMove        = "MOVE"
	eventConsumeFood = "CONSUMEFOOD"
	eventSpawnFood   = "SPAWNFOOD"
	eventExtend      = "EXTEND"
)

func (server *Server) handleExtend(initiatorId string) {
	server.mu.RLock()

	initiatorWorm := server.worms[initiatorId]
	oldWormLength := len(initiatorWorm.positions)
	tailPos := initiatorWorm.positions[oldWormLength-1]

	server.mu.RUnlock()

	newWormPositions := append(initiatorWorm.positions, tailPos)

	server.mu.Lock()

	initiatorWorm.foodConsumed = 0
	initiatorWorm.foodNeeded = (oldWormLength) + 1*server.levelMultiplier
	initiatorWorm.positions = newWormPositions

	server.mu.Unlock()

	server.broadcast([]byte(eventExtend + "\n" + initiatorId + "," + positionToString(&tailPos)))
}

func (server *Server) handleConsumeFood(initiatorId string, headPosCell *cellInfo, headPos *pos) {
	server.mu.Lock()

	initiatorWorm := server.worms[initiatorId]
	headPosCell.food = false
	initiatorWorm.foodConsumed++

	server.mu.Unlock()

	if initiatorWorm.foodConsumed == initiatorWorm.foodNeeded {
		server.handleExtend(initiatorId)
	}

	msg := eventConsumeFood + "\n" + initiatorId + "," + positionToString(headPos)
	server.broadcast([]byte(msg))
}

func (server *Server) handleMove(initiator *websocket.Conn, initiatorId string, dir string) {
	server.move(initiatorId, dir)

	server.mu.RLock()

	initiatorWorm := server.worms[initiatorId]

	headPos := &initiatorWorm.positions[0]
	headPosCell := &server.grid[headPos.x][headPos.y]

	server.mu.RUnlock()

	if headPosCell.food {
		server.handleConsumeFood(initiatorId, headPosCell, headPos)
	}

	server.broadcastExcept([]byte(eventMove+"\n"+initiatorId+","+dir), initiator)
}

func (server *Server) handleInit(initiator *websocket.Conn, initiatorId string) {
	msg := eventInit + "\n"

	newWormMsg := initiatorId + "," + positionsToString(server.worms[initiatorId].positions)
	msg += newWormMsg

	existingWormsMsg := ""

	for id, worm := range server.worms {
		if id != initiatorId {
			existingWormsMsg += id + "," + positionsToString(worm.positions)
		}
	}

	if existingWormsMsg != "" {
		msg += "|" + existingWormsMsg[0:len(existingWormsMsg)-1]
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
		msg += "|" + foodPositionsMsg[0:len(foodPositionsMsg)-1]
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
		case eventMove:
			{
				server.handleMove(ws, id, data)
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
