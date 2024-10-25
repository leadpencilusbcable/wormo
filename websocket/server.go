package websocket

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

const (
	wormMoveInterval             = 500 * time.Millisecond
	foodInterval                 = 5 * time.Second
	maxFoodPerInterval           = 5
	bombInterval                 = 4 * time.Second
	maxBombRadius                = 3
	minBombRadius                = 1
	minBombDetonationTimeSeconds = 5
	maxBombDetonationTimeSeconds = 12
)

type bomb struct {
	bombPosition     pos
	positions        []pos
	timeToDetonation int
}

type Server struct {
	gridWidth       int
	gridHeight      int
	levelMultiplier int
	wormIdCounter   uint64
	bombIdCounter   uint64
	worms           map[string]*worm
	wormConns       map[*websocket.Conn]string
	bombs           map[string]*bomb
	grid            [][]cellInfo
	Server          *http.Server
	mu              sync.RWMutex
}

type collisionInfo struct {
	loss  bool
	gains int
}

func (server *Server) broadcast(msg []byte) {
	server.mu.RLock()

	for k := range server.wormConns {
		k.Write(msg)
	}

	server.mu.RUnlock()
}

func (server *Server) broadcastExcept(msg []byte, except *websocket.Conn) {
	server.mu.RLock()

	for k := range server.wormConns {
		if k != except {
			k.Write(msg)
		}
	}

	server.mu.RUnlock()
}

func (server *Server) handleBomb() {
	server.bombIdCounter++
	bombId := server.bombIdCounter
	bombIdStr := strconv.FormatUint(bombId, 10)

	radius := rand.Intn(maxBombRadius-minBombRadius) + minBombRadius

	x := rand.Intn(server.gridWidth)
	y := rand.Intn(server.gridHeight)

	lowX := int(math.Max(0, float64(x-radius)))
	highX := int(math.Min(float64(server.gridWidth-1), float64(x+radius)))

	lowY := int(math.Max(0, float64(y-radius)))
	highY := int(math.Min(float64(server.gridHeight-1), float64(y+radius)))

	width := highX - lowX + 1
	height := highY - lowY + 1

	bombPositions := make([]pos, width*height)

	index := 0

	for curX := lowX; curX <= highX; curX++ {
		for curY := lowY; curY <= highY; curY++ {
			bombPositions[index] = pos{curX, curY}
			index++
		}
	}

	detonationTimeSeconds := rand.Intn(maxBombDetonationTimeSeconds-minBombDetonationTimeSeconds) + minBombDetonationTimeSeconds
	detonationTime := time.Duration(detonationTimeSeconds) * time.Second

	server.mu.Lock()
	server.bombs[bombIdStr] = &bomb{pos{x, y}, bombPositions, detonationTimeSeconds}
	server.mu.Unlock()

	server.broadcast([]byte(eventSpawnBomb + "\n" + bombIdStr + "|" + strconv.FormatInt(int64(detonationTimeSeconds), 10) + "|" + positionToString(&pos{x, y}) + "|" + positionsToString(bombPositions)))

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		bomb := server.bombs[bombIdStr]

		for range ticker.C {
			server.mu.Lock()
			bomb.timeToDetonation--
			server.mu.Unlock()

			if bomb.timeToDetonation <= 0 {
				return
			}
		}
	}()

	time.Sleep(detonationTime)

	damageMap := map[string]int{}

	server.mu.RLock()

	for _, pos := range bombPositions {
		worm := server.grid[pos.x][pos.y].worm

		if worm != "" {
			damageMap[worm]++
		}
	}

	server.mu.RUnlock()

	wormsMsg := ""

	for wormId, damage := range damageMap {
		worm := server.worms[wormId]

		server.reduce(worm, damage)
		wormsMsg += wormId + "," + positionsToString(worm.positions) + "\n"
	}

	server.mu.Lock()
	delete(server.bombs, bombIdStr)
	server.mu.Unlock()

	if wormsMsg != "" {
		server.broadcast([]byte(eventDetonateBomb + "\n" + bombIdStr + "|" + wormsMsg[:len(wormsMsg)-1]))
	} else {
		server.broadcast([]byte(eventDetonateBomb + "\n" + bombIdStr))
	}
}

func (server *Server) startBombSpawn() {
	ticker := time.NewTicker(bombInterval)

	defer ticker.Stop()

	for range ticker.C {
		server.mu.RLock()
		unlocked := false

		if len(server.wormConns) > 0 {
			server.mu.RUnlock()
			unlocked = true

			go server.handleBomb()
		}

		if !unlocked {
			server.mu.RUnlock()
		}
	}
}

func (server *Server) newFood() *pos {
	for i := 0; i < 5; i++ {
		x := rand.Intn(server.gridWidth)
		y := rand.Intn(server.gridHeight)

		server.mu.RLock()

		food := &server.grid[x][y].food

		if *food {
			server.mu.RUnlock()
			continue
		}

		server.mu.RUnlock()

		server.mu.Lock()
		*food = true
		server.mu.Unlock()

		return &pos{x, y}
	}

	return nil
}

func (server *Server) startFoodSpawn() {
	ticker := time.NewTicker(foodInterval)

	defer ticker.Stop()

	for range ticker.C {
		server.mu.RLock()
		unlocked := false

		if len(server.wormConns) > 0 {
			server.mu.RUnlock()
			unlocked = true

			newFoodPos := server.newFood()

			foodStr := ""

			if newFoodPos != nil {
				foodStr += positionToString(newFoodPos)
			}

			for i := 0; i < rand.Intn(maxFoodPerInterval); i++ {
				newFoodPos = server.newFood()

				if newFoodPos != nil {
					if foodStr != "" {
						foodStr += ","
					}

					foodStr += positionToString(newFoodPos)
				}
			}

			if newFoodPos != nil {
				server.broadcast([]byte(eventSpawnFood + "\n" + foodStr))
			}
		}

		if !unlocked {
			server.mu.RUnlock()
		}
	}
}

func (server *Server) moveWorms() {
	ticker := time.NewTicker(wormMoveInterval)

	defer ticker.Stop()

	for range ticker.C {
		collisions := map[string]*collisionInfo{}

		server.mu.RLock()
		unlocked := false

		if len(server.wormConns) > 0 {
			server.mu.RUnlock()
			unlocked = true

			msg := eventMove + "\n"

			for id, worm := range server.worms {
				server.move(id, worm.direction, &collisions)
			}

			wormsMsg := ""

			for id, worm := range server.worms {
				collison, didCollide := collisions[id]

				if didCollide {
					if collison.loss {
						server.reduce(worm, len(worm.positions)/2)
					}
					if collison.gains > 0 {
						server.extend(worm, collison.gains)
					}
				}

				wormsMsg += id + "," + positionsToString(worm.positions) + "\n"
			}

			msg += wormsMsg[:len(wormsMsg)-1]

			server.broadcast([]byte(msg))
		}

		if !unlocked {
			server.mu.RUnlock()
		}
	}
}

func (server *Server) initGrid() {
	server.grid = make([][]cellInfo, server.gridWidth)

	for i := range server.grid {
		server.grid[i] = make([]cellInfo, server.gridHeight)
	}

	for x := 0; x < server.gridWidth; x++ {
		for y := 0; y < server.gridHeight; y++ {
			server.grid[x][y] = cellInfo{
				"",
				false,
			}
		}
	}
}

func NewServer(port uint16, gridWidth uint8, gridHeight uint8, levelMultiplier uint8) *Server {
	wsServer := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	server := &Server{
		int(gridWidth),
		int(gridHeight),
		int(levelMultiplier),
		0,
		0,
		map[string]*worm{},
		map[*websocket.Conn]string{},
		map[string]*bomb{},
		[][]cellInfo{},
		wsServer,
		sync.RWMutex{},
	}

	wsMux := http.NewServeMux()
	wsMux.Handle("/", websocket.Handler(server.handle))

	wsServer.Handler = wsMux

	server.initGrid()

	go server.startFoodSpawn()
	go server.startBombSpawn()
	go server.moveWorms()

	return server
}
