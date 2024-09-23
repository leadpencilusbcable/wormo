package http

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var gameFile []byte = nil
var errorFile []byte = nil
var notFoundFile []byte = nil

func serveDirectory(dir string, w *http.ResponseWriter, r *http.Request) error {
	uriBlocks := strings.SplitAfter(r.RequestURI, "/")
	filePath := uriBlocks[len(uriBlocks)-1]

	if strings.Contains(filePath, "..") {
		return errors.New("cannot go upwards when looking for file")
	}

	filePath = "./" + dir + "/" + filePath

	file, error := os.ReadFile(filePath)

	writer := (*w)

	if error != nil {
		log.Println("Error reading " + filePath)

		writer.WriteHeader(500)
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Write(notFoundFile)

		return nil
	}

	mimeType := "any"

	if strings.HasSuffix(filePath, ".js") {
		mimeType = "text/javascript"
	}

	writer.Header().Set("Content-Type", mimeType+"; charset=utf-8")
	writer.Write(file)

	return nil
}

func createGameFile(x int, y int) error {
	template, error := template.New("game.html").Funcs(template.FuncMap{
		"loop": func(n int) []struct{} {
			return make([]struct{}, n)
		},
	}).ParseFiles("./templates/game.html")

	if error != nil {
		return error
	}

	gameFile, error := os.Create("./public/pages/game.html")

	if error != nil {
		return error
	}

	defer gameFile.Close()

	error = template.Execute(gameFile, struct {
		X         int
		Y         int
		TotalSize int
	}{X: x, Y: y, TotalSize: x * y})

	if error != nil {
		return error
	}

	return nil
}

func handleStyles(w http.ResponseWriter, r *http.Request) {
	serveDirectory("public/styles", &w, r)
}

func handleScripts(w http.ResponseWriter, r *http.Request) {
	serveDirectory("public/scripts", &w, r)
}

func handle(w http.ResponseWriter, r *http.Request) {
	if gameFile == nil {
		var error error

		gameFile, error = os.ReadFile("./public/pages/game.html")

		if error != nil {
			log.Println("Error reading game file")

			w.WriteHeader(500)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(errorFile)

			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(gameFile)
}

func StartServer(port uint, x int, y int) error {
	error := createGameFile(x, y)

	if error != nil {
		return error
	}

	errorFile, error = os.ReadFile("./public/pages/error.html")

	if error != nil {
		return error
	}

	notFoundFile, error = os.ReadFile("./public/pages/pagenotfound.html")

	if error != nil {
		return error
	}

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", handle)
	httpMux.HandleFunc("/scripts/", handleScripts)
	httpMux.HandleFunc("/styles/", handleStyles)

	httpServer := &http.Server{
		Addr:    ":" + strconv.FormatUint(uint64(port), 10),
		Handler: httpMux,
	}

	httpServer.ListenAndServe()

	return nil
}
