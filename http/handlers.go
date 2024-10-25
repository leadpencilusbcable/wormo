package http

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func serveDirectory(dir string, w *http.ResponseWriter, r *http.Request, notFoundFile *[]byte) {
	slashCount := 0
	secondSlash := -1

	for i := 0; i < len(r.RequestURI); i++ {
		if r.RequestURI[i] == '/' {
			slashCount++

			if slashCount == 2 {
				secondSlash = i
				break
			}
		}
	}

	writer := (*w)

	if secondSlash == -1 {
		log.Println("URL malformed for directory search")

		writer.WriteHeader(500)
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Write(*notFoundFile)

		return
	}

	filePath := dir + r.RequestURI[secondSlash:]

	file, error := os.ReadFile(filePath)

	if error != nil {
		log.Println("Error reading " + filePath)

		writer.WriteHeader(500)
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Write(*notFoundFile)

		return
	}

	extension := filePath[strings.LastIndexByte(filePath, '.')+1:]

	var mimeType string

	switch extension {
	case "js":
		mimeType = "text/javascript"
	case "css":
		mimeType = "text/css"
	case "png":
		mimeType = "image/png"
	default:
		mimeType = "any"
	}

	writer.Header().Set("Content-Type", mimeType+"; charset=utf-8")
	writer.Write(file)
}

func (server *Server) handleStyles(w http.ResponseWriter, r *http.Request) {
	serveDirectory(server.stylesPath, &w, r, &server.notFoundFile)
}

func (server *Server) handleScripts(w http.ResponseWriter, r *http.Request) {
	serveDirectory(server.scriptsPath, &w, r, &server.notFoundFile)
}

func (server *Server) handleImages(w http.ResponseWriter, r *http.Request) {
	serveDirectory(server.imagesPath, &w, r, &server.notFoundFile)
}

func (server *Server) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(server.gameFile)
}
