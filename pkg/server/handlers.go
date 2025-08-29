package server

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"webnote/pkg/storage"
	"webnote/pkg/websocket"

	"github.com/gorilla/mux"
)

func (s *Server) noteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := vars["path"]
		file, size, err := storage.GetNote(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if file == nil {
			data := struct {
				Path    string
				Content string
			}{
				Path:    path,
				Content: "",
			}
			if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		defer file.Close()

		userAgent := r.Header.Get("User-Agent")
		if _, raw := r.URL.Query()["raw"]; raw || strings.HasPrefix(userAgent, "curl") || strings.HasPrefix(userAgent, "Wget") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			io.Copy(w, file)
			return
		}

		if size > s.maxContentSize {
			http.Error(w, "Note is too large to display.", http.StatusRequestEntityTooLarge)
			return
		}
		content, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Path    string
			Content string
		}{
			Path:    path,
			Content: string(content),
		}

		if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) saveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := vars["path"]

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form.", http.StatusInternalServerError)
			return
		}
		content := r.FormValue("content")

		// Add check for content size before saving
		if int64(len(content)) > s.maxContentSize {
			http.Error(w, "Note is too large to save. Maximum size is "+strconv.FormatInt(s.maxContentSize/1024, 10)+"KB.", http.StatusRequestEntityTooLarge)
			return
		}

		var contentLength int64
		if len(strings.TrimSpace(content)) == 0 {
			contentLength = 0
		} else {
			contentLength = int64(len(content))
		}

		err := storage.SaveNote(path, strings.NewReader(content), contentLength)
		if err != nil {
			if err == storage.ErrStorageTooLarge {
				http.Error(w, "Storage is overloaded.", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		broadcastMsg := &websocket.BroadcastMsg{
			Path:    path,
			Content: []byte(content),
		}
		s.hub.Broadcast <- broadcastMsg

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Note saved.\n"))
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/"+randStringBytes(5), http.StatusFound)
}

func storageOverloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/save/") && r.Method == "POST" {
			if r.ContentLength > 0 && storage.IsStorageOverloaded() {
				http.Error(w, "Storage is overloaded.", http.StatusServiceUnavailable)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
