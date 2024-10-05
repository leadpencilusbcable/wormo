package websocket

import (
	"math/rand"
	"strconv"
	"sync/atomic"
)

type pos struct {
	x int
	y int
}

type worm struct {
	positions    []pos
	direction    string
	foodConsumed int
	foodNeeded   int
}

type cellInfo struct {
	worm string
	food bool
}

func (server *Server) extend(worm *worm) {
	server.mu.RLock()

	oldWormLength := len(worm.positions)
	tailPos := worm.positions[oldWormLength-1]

	server.mu.RUnlock()

	newWormPositions := append(worm.positions, tailPos)

	server.mu.Lock()

	worm.foodConsumed = 0
	worm.foodNeeded = (oldWormLength + 1) * server.levelMultiplier
	worm.positions = newWormPositions

	server.mu.Unlock()
}

func (server *Server) consumeFood(id string, headPosCell *cellInfo, headPos *pos) {
	server.mu.Lock()
	worm := server.worms[id]

	headPosCell.food = false
	worm.foodConsumed++

	server.mu.Unlock()

	if worm.foodConsumed == worm.foodNeeded {
		server.extend(worm)
	}

	server.broadcast([]byte(eventConsumeFood + "\n" + id + "," + positionToString(headPos) + "|" + strconv.Itoa(worm.foodConsumed) + "/" + strconv.Itoa(worm.foodNeeded)))
}

func (server *Server) move(id string, dir string) {
	server.mu.RLock()

	worm := server.worms[id]

	positions := worm.positions
	tailPos := positions[len(positions)-1]

	server.mu.RUnlock()

	server.mu.Lock()
	server.grid[tailPos.x][tailPos.y].worm = ""
	server.mu.Unlock()

	for i := len(positions) - 1; i > 0; i-- {
		pos := &positions[i]
		nextPos := positions[i-1]

		*pos = nextPos
	}

	headPos := &positions[0]

	switch dir {
	case "U":
		headPos.y--
	case "D":
		headPos.y++
	case "L":
		headPos.x--
	case "R":
		headPos.x++
	}

	server.mu.Lock()

	worm.positions = positions

	headPosCell := &server.grid[headPos.x][headPos.y]

	//this won't work in cases where worm's old tailpos is still taken up by worm. temp solution
	server.grid[tailPos.x][tailPos.y].worm = ""
	headPosCell.worm = id

	server.mu.Unlock()

	if headPosCell.food {
		server.consumeFood(id, headPosCell, headPos)
	}
}

func positionToString(position *pos) string {
	return strconv.Itoa(position.x) + ":" + strconv.Itoa(position.y)
}

func positionsToString(positions []pos) string {
	if len(positions) == 0 {
		return ""
	}

	positionsString := positionToString(&positions[0])

	for i := 1; i < len(positions); i++ {
		positionsString += "," + positionToString(&positions[i])
	}

	return positionsString
}

func (server *Server) newWorm() string {
	atomic.AddUint64(&server.idCounter, 1)
	id := strconv.FormatUint(server.idCounter, 10)

	x := rand.Intn(server.gridWidth-10) + 5
	y := rand.Intn(server.gridHeight-10) + 5

	wormPos := []pos{{x, y}, {x - 1, y}, {x - 2, y}}
	server.worms[id] = &worm{
		wormPos,
		"L",
		0,
		3 * server.levelMultiplier,
	}

	return id
}
