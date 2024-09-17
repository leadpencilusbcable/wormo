package websocket

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

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
	for i := 0; i < len(worm.pos)-1; i++ {
		pos := &worm.pos[i]
		nextPos := worm.pos[i+1]

		*pos = nextPos
	}

	headPos := &worm.pos[len(worm.pos)-1]

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
}

const INIT_WORM_LENGTH int = 3
const (
	eventInit    = "INIT"
	eventNewWorm = "NEW"
	eventMove    = "MOVE"
)

var cells map[pos]cellInfo = map[pos]cellInfo{}
var conns map[*websocket.Conn]worm = map[*websocket.Conn]worm{}

var ROWS int
var COLS int

var wormCounter = 0

func initGrid(x int, y int) {
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			cells[pos{i, j}] = cellInfo{nil}
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
	x := rand.Intn(COLS-10) + 5
	y := rand.Intn(ROWS-10) + 5

	wormPos := []pos{{x, y}, {x - 1, y}, {x - 2, y}}
	wormCounter++

	return worm{
		strconv.FormatInt(int64(wormCounter), 10),
		wormPos,
	}
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

func StartServer(port uint, x int, y int) {
	ROWS = y
	COLS = x

	wsMux := http.NewServeMux()
	wsMux.Handle("/", websocket.Handler(handle))

	wsServer := &http.Server{
		Addr:    ":" + strconv.FormatUint(uint64(port), 10),
		Handler: wsMux,
	}

	initGrid(x, y)

	wsServer.ListenAndServe()
}
