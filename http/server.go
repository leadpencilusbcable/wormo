package http

import (
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	gameFile     []byte
	errorFile    []byte
	notFoundFile []byte
	stylesPath   string
	scriptsPath  string
	Server       *http.Server
}

func createGameFile(gridWidth uint8, gridHeight uint8, levelMultiplier uint8, gameFilePath string) error {
	template, error := template.New("game.html").Funcs(template.FuncMap{
		"loop": func(n int) []struct{} {
			return make([]struct{}, n)
		},
	}).ParseFiles("./templates/game.html")

	if error != nil {
		return error
	}

	gameFile, error := os.Create(gameFilePath)

	if error != nil {
		return error
	}

	defer gameFile.Close()

	error = template.Execute(gameFile, struct {
		X               int
		Y               int
		TotalSize       int
		LevelMultiplier int
	}{
		int(gridWidth),
		int(gridHeight),
		int(gridWidth) * int(gridHeight),
		int(levelMultiplier),
	})

	if error != nil {
		return error
	}

	return nil
}

func NewServer(
	port uint16,
	gridWidth uint8,
	gridHeight uint8,
	levelMultiplier uint8,
	gameFilePath string,
	errorFilePath string,
	notFoundFilePath string,
	stylesPath string,
	scriptsPath string,
) (*Server, error) {
	error := createGameFile(gridWidth, gridHeight, levelMultiplier, gameFilePath)

	if error != nil {
		return nil, error
	}

	gameFile, error := os.ReadFile(gameFilePath)

	if error != nil {
		return nil, error
	}

	errorFile, error := os.ReadFile(errorFilePath)

	if error != nil {
		return nil, error
	}

	notFoundFile, error := os.ReadFile(notFoundFilePath)

	if error != nil {
		return nil, error
	}

	httpServer := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	server := &Server{
		gameFile,
		errorFile,
		notFoundFile,
		stylesPath,
		scriptsPath,
		httpServer,
	}

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", server.handle)
	httpMux.HandleFunc("/scripts/", server.handleScripts)
	httpMux.HandleFunc("/styles/", server.handleStyles)

	httpServer.Handler = httpMux

	return server, nil
}
