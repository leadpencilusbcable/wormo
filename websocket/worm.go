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

func (server *Server) reduce(worm *worm, amount int) {
	server.mu.RLock()
	newLength := len(worm.positions) - amount
	server.mu.RUnlock()

	if newLength < 1 {
		server.mu.Lock()

		for i := 1; i < len(worm.positions); i++ {
			server.grid[worm.positions[i].x][worm.positions[i].y].worm = ""
		}

		worm.positions = []pos{worm.positions[0]}
		worm.foodConsumed = 0
		worm.foodNeeded = 1

		server.mu.Unlock()
	} else {
		server.mu.Lock()

		newPositions := make([]pos, newLength)

		for i := 0; i < len(worm.positions); i++ {
			if i < newLength {
				newPositions[i] = worm.positions[i]
			} else {
				server.grid[worm.positions[i].x][worm.positions[i].y].worm = ""
			}
		}

		worm.positions = newPositions
		worm.foodConsumed = 0
		worm.foodNeeded = newLength * server.levelMultiplier

		server.mu.Unlock()
	}
}

func (server *Server) extend(worm *worm, amount int) {
	server.mu.RLock()

	oldWormLength := len(worm.positions)
	tailPos := worm.positions[oldWormLength-1]

	server.mu.RUnlock()

	var newWormPositions []pos

	if amount == 1 {
		newWormPositions = append(worm.positions, tailPos)
	} else {
		newWormPositions = make([]pos, oldWormLength+amount)

		for i := range newWormPositions {
			if i < oldWormLength {
				newWormPositions[i] = worm.positions[i]
			} else {
				newWormPositions[i] = tailPos
			}
		}
	}

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
		server.extend(worm, 1)
	}

	server.broadcast([]byte(eventConsumeFood + "\n" + id + "," + positionToString(headPos) + "|" + strconv.Itoa(worm.foodConsumed) + "/" + strconv.Itoa(worm.foodNeeded)))
}

func (server *Server) move(id string, dir string, collisions *map[string]*collisionInfo) {
	server.mu.RLock()

	worm := server.worms[id]

	positions := worm.positions
	tailPos := positions[len(positions)-1]
	headPos := positions[0]

	server.mu.RUnlock()

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

	outOfBounds := headPos.x == -1 || headPos.x == server.gridWidth || headPos.y == -1 || headPos.y == server.gridHeight

	collison, existingCollision := (*collisions)[id]

	if outOfBounds {
		if existingCollision {
			collison.loss = true
		} else {
			(*collisions)[id] = &collisionInfo{
				true,
				0,
			}
		}

		return
	}

	enemyWormId := server.grid[headPos.x][headPos.y].worm

	if enemyWormId != id && enemyWormId != "" {
		server.mu.RLock()
		enemyWorm := server.worms[enemyWormId]
		enemyWormHeadPos := enemyWorm.positions[0]
		server.mu.RUnlock()

		if len(worm.positions) == 1 || (enemyWormHeadPos.x == headPos.x && enemyWormHeadPos.y == headPos.y) {
			return
		}

		if existingCollision {
			collison.loss = true
		} else {
			(*collisions)[id] = &collisionInfo{
				true,
				0,
			}
		}

		dmg := len(positions) / 2
		enemyCollison, existingEnemyCollision := (*collisions)[enemyWormId]

		if existingEnemyCollision {
			enemyCollison.gains += dmg
		} else {
			(*collisions)[enemyWormId] = &collisionInfo{
				false,
				dmg,
			}
		}

		return
	}

	tailPosOverlap := false

	for i := len(positions) - 1; i > 0; i-- {
		pos := &positions[i]
		nextPos := positions[i-1]

		if nextPos.x == tailPos.x && nextPos.y == tailPos.y {
			tailPosOverlap = true
		}

		server.mu.Lock()
		*pos = nextPos
		server.mu.Unlock()
	}

	server.mu.Lock()

	worm.positions = positions
	worm.positions[0] = headPos

	headPosCell := &server.grid[headPos.x][headPos.y]

	if !tailPosOverlap {
		server.grid[tailPos.x][tailPos.y].worm = ""
	}

	headPosCell.worm = id

	server.mu.Unlock()

	if headPosCell.food {
		server.consumeFood(id, headPosCell, &headPos)
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
		positions:    wormPos,
		direction:    "R",
		foodConsumed: 0,
		foodNeeded:   3 * server.levelMultiplier,
	}

	return id
}
