package server

import (
	"html/template"
	"net/http"
	"webnote/pkg/storage"
	"webnote/pkg/websocket"

	"github.com/gorilla/mux"
)

type Server struct {
	router    *mux.Router
	templates *template.Template
	hub       *websocket.Hub
}

func NewServer(hub *websocket.Hub) *Server {
	s := &Server{
		router: mux.NewRouter(),
		hub:    hub,
	}
	s.templates = template.Must(template.ParseFiles("index.html"))
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
	s.router.HandleFunc("/{path}", s.noteHandler())
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
