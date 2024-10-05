package websocket

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

const wormMoveInterval = 500 * time.Millisecond

const foodInterval = 5 * time.Second
const maxFoodPerInterval = 5

type Server struct {
	gridWidth       int
	gridHeight      int
	levelMultiplier int
	idCounter       uint64
	worms           map[string]*worm
	wormConns       map[*websocket.Conn]string
	grid            [][]cellInfo
	Server          *http.Server
	mu              sync.RWMutex
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
		server.mu.RLock()
		unlocked := false

		if len(server.wormConns) > 0 {
			server.mu.RUnlock()
			unlocked = true

			msg := eventMove + "\n"

			wormsMsg := ""

			for ws, id := range server.wormConns {
				worm := server.worms[id]

				server.move(ws, worm.direction)
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
		map[string]*worm{},
		map[*websocket.Conn]string{},
		[][]cellInfo{},
		wsServer,
		sync.RWMutex{},
	}

	wsMux := http.NewServeMux()
	wsMux.Handle("/", websocket.Handler(server.handle))

	wsServer.Handler = wsMux

	server.initGrid()

	go server.startFoodSpawn()
	go server.moveWorms()

	return server
}
