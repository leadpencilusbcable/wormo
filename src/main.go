package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type GameData struct {
	X         int
	Y         int
	TotalSize int
}

var gameFile []byte = nil
var errorFile []byte = nil
var notFoundFile []byte = nil

func serveDirectory(dir string, w *http.ResponseWriter, r *http.Request) error {
	filePath := strings.SplitAfter(r.RequestURI, "/"+dir+"/")[1]

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

	writer.Header().Set("Content-Type", "any; charset=utf-8")
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

	gameFile, error := os.Create("./pages/game.html")

	if error != nil {
		return error
	}

	defer gameFile.Close()

	error = template.Execute(gameFile, GameData{x, y, x * y})

	if error != nil {
		return error
	}

	return nil
}

func handleStyles(w http.ResponseWriter, r *http.Request) {
	serveDirectory("styles", &w, r)
}

func handleScripts(w http.ResponseWriter, r *http.Request) {
	serveDirectory("scripts", &w, r)
}

func handle(w http.ResponseWriter, r *http.Request) {
	if gameFile == nil {
		var error error

		gameFile, error = os.ReadFile("./pages/game.html")

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

func initialise() error {
	error := createGameFile(40, 30)

	if error != nil {
		return error
	}

	errorFile, error = os.ReadFile("./pages/error.html")

	if error != nil {
		return error
	}

	notFoundFile, error = os.ReadFile("./pages/pagenotfound.html")

	if error != nil {
		return error
	}

	return nil
}

func main() {
	error := initialise()

	if error != nil {
		log.Panic(error)
	}

	http.HandleFunc("/", handle)
	http.HandleFunc("/scripts/", handleScripts)
	http.HandleFunc("/styles/", handleStyles)
	http.ListenAndServe(":8000", nil)
}
