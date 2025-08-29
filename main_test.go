package main

import (
	"bytes"
	"html/template"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

var testRouter http.Handler

func setup() {
	// Reset storage for a clean test environment.
	os.RemoveAll(savePath)
	initStorage()
	templates = template.Must(template.ParseFiles("index.html"))

	r := mux.NewRouter()
	r.Use(storageOverloadMiddleware)
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/save/{path}", saveHandler).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/{path}", noteHandler)
	testRouter = r
}

func TestMain(m *testing.M) {
	// Setup can be done here if it's for all tests in the package.
	// For benchmark, it's better to control setup within the benchmark function
	// to not include setup time in the benchmark measurement.
	os.Exit(m.Run())
}

func BenchmarkSaveHandler(b *testing.B) {
	setup()
	// Prepare some data to save.
	noteData := []byte("This is a test note for benchmarking.")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			path := randStringBytes(10)
			req := httptest.NewRequest("POST", "/save/"+path, bytes.NewReader(noteData))
			req.RemoteAddr = "192.0.2.1:" + strconv.Itoa(rand.Intn(10000))
			req.ContentLength = int64(len(noteData))
			rr := httptest.NewRecorder()
			testRouter.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Errorf("saveHandler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}
		}
	})
}

func BenchmarkNoteHandler(b *testing.B) {
	setup()
	// First, save a note to be read.
	path := "benchmark_note"
	noteData := []byte("This is a permanent note for read benchmark.")
	err := saveNote(path, bytes.NewReader(noteData), int64(len(noteData)))
	if err != nil {
		b.Fatalf("Failed to save a note for reading benchmark: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/"+path, nil)
			rr := httptest.NewRecorder()
			testRouter.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Errorf("noteHandler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}
		}
	})
}
