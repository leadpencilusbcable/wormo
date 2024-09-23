package websocket

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

type pos struct {
	x int
	y int
}

type worm struct {
	id  string
	pos []pos
}

func (worm worm) move(dir string) {
	for i := len(worm.pos) - 1; i > 0; i-- {
		pos := &worm.pos[i]
		nextPos := worm.pos[i-1]

		*pos = nextPos
	}

	headPos := &worm.pos[0]

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
}

type cellInfo struct {
	worm *websocket.Conn
	food bool
}

const INIT_WORM_LENGTH int = 3

const (
	eventInit      = "INIT"
	eventNewWorm   = "NEW"
	eventMove      = "MOVE"
	eventSpawnFood = "SPAWNFOOD"
	eventExtend    = "EXTEND"
)

var cells map[pos]cellInfo = map[pos]cellInfo{}
var conns map[*websocket.Conn]worm = map[*websocket.Conn]worm{}

var rows int
var cols int

var wormCounter = 0

func initGrid(x int, y int) {
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			cells[pos{i, j}] = cellInfo{nil, false}
		}
	}
}

func wormToString(worm worm) string {
	ret := worm.id + "," + strconv.Itoa(worm.pos[0].x) + ":" + strconv.Itoa(worm.pos[0].y)

	for i := 1; i < len(worm.pos); i++ {
		ret += "," + strconv.Itoa(worm.pos[i].x) + ":" + strconv.Itoa(worm.pos[i].y)
	}

	return ret
}

func makeWorm() worm {
	x := rand.Intn(cols-10) + 5
	y := rand.Intn(rows-10) + 5

	wormPos := []pos{{x, y}, {x - 1, y}, {x - 2, y}}
	wormCounter++

	return worm{
		strconv.FormatInt(int64(wormCounter), 10),
		wormPos,
	}
}

func newFood() *pos {
	for i := 0; i < 10; i++ {
		x := rand.Intn(cols)
		y := rand.Intn(rows)

		foodPos := pos{x, y}

		if cells[foodPos].food {
			continue
		}

		return &foodPos
	}

	return nil
}

func handle(ws *websocket.Conn) {
	log.Println("Incoming connection: ", ws.LocalAddr())

	conns[ws] = makeWorm()

	readFromConnection(ws)
}

func readFromConnection(ws *websocket.Conn) {
	buffer := make([]byte, 1024)

	for {
		length, error := ws.Read(buffer)

		if error != nil {
			if error == io.EOF {
				delete(conns, ws)
				break
			}

			log.Println("Read error: ", error)
			continue
		}

		msg := string(buffer[:length])
		log.Println("Msg: ", msg)

		event := msg
		data := msg

		eventMessageSplit := strings.Index(msg, "\n")

		if eventMessageSplit != -1 {
			event = msg[0:eventMessageSplit]
			data = msg[eventMessageSplit+1 : length]
		}

		switch event {
		case eventMove:
			{
				conns[ws].move(data)
				broadcast([]byte(eventMove+"\n"+conns[ws].id+","+data), ws)
			}
		case eventExtend:
			{
				strPos := strings.Split(data, ":")

				x, err := strconv.Atoi(strPos[0])

				if err != nil {
					log.Println("invalid x pos received for extend")
					break
				}

				y, err := strconv.Atoi(strPos[1])

				if err != nil {
					log.Println("invalid y pos received for extend")
					break
				}

				tempPos := conns[ws].pos
				conns[ws] = worm{
					conns[ws].id,
					append(tempPos, pos{x, y}),
				}

				broadcast([]byte(eventExtend+"\n"+conns[ws].id+","+data), ws)
			}
		case eventInit:
			{
				newWormStr := wormToString(conns[ws])
				existingWormsStr := newWormStr

				for k, v := range conns {
					if k != ws {
						existingWormsStr += "\n" + wormToString(v)
					}
				}

				ws.Write([]byte(existingWormsStr))
				broadcast([]byte(eventNewWorm+"\n"+newWormStr), ws)
			}
		}
	}
}

func broadcast(msg []byte, except *websocket.Conn) {
	if except != nil {
		for k := range conns {
			if k != except {
				k.Write(msg)
			}
		}
	} else {
		for k := range conns {
			k.Write(msg)
		}
	}
}

const foodInterval = 5 * time.Second

func StartServer(port uint, x int, y int) {
	rows = y
	cols = x

	wsMux := http.NewServeMux()
	wsMux.Handle("/", websocket.Handler(handle))

	wsServer := &http.Server{
		Addr:    ":" + strconv.FormatUint(uint64(port), 10),
		Handler: wsMux,
	}

	initGrid(x, y)

	ticker := time.NewTicker(foodInterval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if len(conns) > 0 {
				newFoodPos := newFood()

				if newFoodPos != nil {
					broadcast([]byte(eventSpawnFood+"\n"+strconv.Itoa(newFoodPos.x)+":"+strconv.Itoa(newFoodPos.y)), nil)
				}
			}
		}
	}()

	wsServer.ListenAndServe()
}
