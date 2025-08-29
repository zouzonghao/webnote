package server

import (
	"html/template"
	"net/http"
	"strconv"
	"webnote/pkg/storage"
	"webnote/pkg/websocket"

	"github.com/gorilla/mux"
)

type Server struct {
	router         *mux.Router
	templates      *template.Template
	hub            *websocket.Hub
	maxContentSize int64
}

func NewServer(hub *websocket.Hub, maxContentSize int64) *Server {
	s := &Server{
		router:         mux.NewRouter().StrictSlash(true),
		hub:            hub,
		maxContentSize: maxContentSize,
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	s.templates = template.Must(template.New("").Funcs(funcMap).ParseFiles("index.html", "history.html"))
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.router.Use(storageOverloadMiddleware)

	s.router.HandleFunc("/", rootHandler)
	s.router.HandleFunc("/save/{path}", s.saveHandler()).Methods("POST")
	s.router.HandleFunc("/ws/{path}", s.serveWsHandler())
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.router.HandleFunc("/{path}/{version:[0-9]+}", s.historyHandler()).Methods("GET")
	s.router.HandleFunc("/{path:[^/]+}", s.noteHandler())
}

func (s *Server) serveWsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path validation is now part of the handler logic
		vars := mux.Vars(r)
		path := vars["path"]
		if !storage.IsValidPath(path) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		websocket.ServeWs(s.hub, w, r)
	}
}

func (s *Server) historyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := vars["path"]
		versionStr := vars["version"]

		if !storage.IsValidPath(path) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			http.Error(w, "Invalid version", http.StatusBadRequest)
			return
		}

		content, totalVersions, err := storage.GetNoteVersion(path, version)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var prevVersion, nextVersion int
		if version > 1 {
			prevVersion = version - 1
		}
		if version < totalVersions-1 {
			nextVersion = version + 1
		}

		data := struct {
			Path           string
			Content        string
			CurrentVersion int
			PrevVersion    int
			NextVersion    int
			TotalVersions  int
		}{
			Path:           path,
			Content:        string(content),
			CurrentVersion: version,
			PrevVersion:    prevVersion,
			NextVersion:    nextVersion,
			TotalVersions:  totalVersions,
		}

		if err := s.templates.ExecuteTemplate(w, "history.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
