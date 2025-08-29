package main

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

// --- Configuration ---
const (
	// savePath 是用于存储笔记文件的目录名称。
	savePath = "notes"
)

var (
	// maxStorageSize 是所有笔记文件允许占用的最大总存储空间（以字节为单位）。
	maxStorageSize int64 = 10 * 1024 * 1024 // 10MB
	// maxContentSize 是单个笔记内容允许的最大长度（以字节为单位）。
	maxContentSize int64 = 100 * 1024 // 100KB
	// letterBytes 是用于生成随机URL路径的字符集。
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// defaultPort 是当没有通过环境变量 PORT 指定端口时，服务器监听的默认端口。
	defaultPort = "8080"
	// cleanupInterval 是后台清理任务运行的频率，用于清理过期的访问者记录。
	cleanupInterval = 10 * time.Minute
	// visitorTTL 是一个访问者记录在被视为过期并可以被清理之前，可以保持不活动状态的最长时间。
	visitorTTL = 6 * time.Hour
	// rateLimit 是速率限制器允许每秒生成的令牌数。
	rateLimit = 5.0
	// rateBurst 是令牌桶的最大容量。
	rateBurst = 10.0
)

var ErrStorageTooLarge = errors.New("storage size exceeds the limit")

// --- Application State ---
type visitor struct {
	tokens   float64
	lastSeen time.Time
}

var (
	visitors           = make(map[string]*visitor)
	mu                 sync.Mutex
	currentStorageSize int64
	templates          *template.Template
)

// --- Storage Logic ---

func initStorage() {
	if err := os.MkdirAll(savePath, 0755); err != nil {
		log.Fatalf("failed to create storage directory: %v", err)
	}

	var totalSize int64
	err := filepath.Walk(savePath, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to calculate initial storage size: %v", err)
	}
	atomic.StoreInt64(&currentStorageSize, totalSize)
}

func isValidPath(path string) bool {
	return !strings.Contains(path, "..") && !strings.Contains(path, "/")
}

func getNote(path string) (*os.File, int64, error) {
	if !isValidPath(path) {
		return nil, 0, errors.New("invalid path")
	}
	filePath := filepath.Join(savePath, path)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil // Use nil to indicate not found
		}
		return nil, 0, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	return file, info.Size(), nil
}

func saveNote(path string, contentReader io.Reader, contentLength int64) error {
	if !isValidPath(path) {
		return errors.New("invalid path")
	}

	filePath := filepath.Join(savePath, path)

	var oldSize int64
	info, err := os.Stat(filePath)
	if err == nil {
		oldSize = info.Size()
	} else if !os.IsNotExist(err) {
		return err
	}

	// Handle deletion if content length is 0
	if contentLength == 0 {
		if oldSize > 0 {
			if err := os.Remove(filePath); err != nil {
				return err
			}
			atomic.AddInt64(&currentStorageSize, -oldSize)
		}
		return nil
	}

	// Handle write/update
	newSize := contentLength
	sizeDelta := newSize - oldSize

	if atomic.LoadInt64(&currentStorageSize)+sizeDelta > maxStorageSize {
		return ErrStorageTooLarge
	}

	// Create a temporary file to stream the content to
	tempFile, err := os.CreateTemp(savePath, "temp-note-")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file

	// Stream content to the temporary file
	written, err := io.Copy(tempFile, contentReader)
	if err != nil {
		tempFile.Close()
		return err
	}
	tempFile.Close()

	// This is a fallback check in case ContentLength was deceptive
	if written != newSize {
		return errors.New("content length mismatch")
	}

	// Rename the temporary file to the final destination
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return err
	}

	atomic.AddInt64(&currentStorageSize, sizeDelta)
	return nil
}

func isStorageOverloaded() bool {
	return atomic.LoadInt64(&currentStorageSize) > maxStorageSize
}

// --- HTTP Logic ---

func allowVisitor(ip string) bool {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// First time visitor gets a full burst and we take one token.
		visitors[ip] = &visitor{tokens: rateBurst - 1, lastSeen: time.Now()}
		return true
	}

	// Refill tokens based on time passed.
	now := time.Now()
	elapsed := now.Sub(v.lastSeen)
	v.lastSeen = now
	v.tokens += elapsed.Seconds() * rateLimit
	if v.tokens > rateBurst {
		v.tokens = rateBurst
	}

	// Check if there are enough tokens.
	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

func cleanupVisitors() {
	for {
		time.Sleep(cleanupInterval)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > visitorTTL {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func storageOverloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/save/") && r.Method == "POST" {
			// For streaming, we check the ContentLength header.
			// A check for empty content is done in the handler.
			if r.ContentLength > 0 && isStorageOverloaded() {
				http.Error(w, "Storage is overloaded.", http.StatusServiceUnavailable)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if !allowVisitor(r.RemoteAddr) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	http.Redirect(w, r, "/"+randStringBytes(5), http.StatusFound)
}

func noteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	file, size, err := getNote(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if file == nil {
		http.NotFound(w, r)
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

	// For HTML template, we must read the content into memory.
	// This is a trade-off for template rendering.
	if size > maxContentSize { // Safety check
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

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	if !allowVisitor(r.RemoteAddr) {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	vars := mux.Vars(r)
	path := vars["path"]

	// Check content length against the limit
	if r.ContentLength > maxContentSize {
		http.Error(w, "Content exceeds the maximum allowed size.", http.StatusRequestEntityTooLarge)
		return
	}

	// Use a LimitReader to protect against clients sending more data than declared.
	limitedReader := &io.LimitedReader{R: r.Body, N: maxContentSize + 1}
	defer r.Body.Close()

	err := saveNote(path, limitedReader, r.ContentLength)
	if err != nil {
		if err == ErrStorageTooLarge {
			http.Error(w, "Storage is overloaded.", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the reader was read past the limit
	if limitedReader.N <= 0 {
		http.Error(w, "Content exceeds the maximum allowed size.", http.StatusRequestEntityTooLarge)
		// Note: At this point, a partial file might have been saved. A more robust
		// implementation would handle cleaning this up.
		return
	}

	w.WriteHeader(http.StatusOK)
	// 从环境变量加载配置
	if maxSizeStr := os.Getenv("MAX_STORAGE_SIZE"); maxSizeStr != "" {
		if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxStorageSize = size
			log.Printf("Set MAX_STORAGE_SIZE to %d bytes", maxStorageSize)
		} else {
			log.Printf("Warning: could not parse MAX_STORAGE_SIZE env var: %v", err)
		}
	}
	if maxContentStr := os.Getenv("MAX_CONTENT_SIZE"); maxContentStr != "" {
		if size, err := strconv.ParseInt(maxContentStr, 10, 64); err == nil {
			maxContentSize = size
			log.Printf("Set MAX_CONTENT_SIZE to %d bytes", maxContentSize)
		} else {
			log.Printf("Warning: could not parse MAX_CONTENT_SIZE env var: %v", err)
		}
	}
}

func main() {
	initStorage()
	templates = template.Must(template.ParseFiles("index.html"))
	go cleanupVisitors()

	r := mux.NewRouter()
	r.Use(storageOverloadMiddleware)

	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/save/{path}", saveHandler).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/{path}", noteHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("Listening on :%s...", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}
