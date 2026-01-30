package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Celedhrim/btcleaner/internal/cleaner"
	"github.com/Celedhrim/btcleaner/internal/logger"
	"github.com/Celedhrim/btcleaner/internal/transmission"
	"github.com/gorilla/websocket"
)

// Server represents the web server
type Server struct {
	port        int
	webRoot     string
	cleaner     *cleaner.Cleaner
	client      *transmission.Client
	logger      *logger.Logger
	srv         *http.Server
	upgrader    websocket.Upgrader
	logClients  map[*websocket.Conn]bool
	clientMutex sync.RWMutex
}

// New creates a new server instance
func New(port int, webRoot string, cleaner *cleaner.Cleaner, client *transmission.Client, log *logger.Logger) *Server {
	if !strings.HasPrefix(webRoot, "/") {
		webRoot = "/" + webRoot
	}
	if !strings.HasSuffix(webRoot, "/") {
		webRoot = webRoot + "/"
	}

	return &Server{
		port:       port,
		webRoot:    webRoot,
		cleaner:    cleaner,
		client:     client,
		logger:     log,
		upgrader:   websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		logClients: make(map[*websocket.Conn]bool),
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Strip webroot prefix for all handlers
	stripPrefix := s.webRoot
	if stripPrefix == "/" {
		stripPrefix = ""
	} else {
		stripPrefix = strings.TrimSuffix(stripPrefix, "/")
	}

	// API endpoints
	mux.HandleFunc(stripPrefix+"/api/stats", s.handleStats)
	mux.HandleFunc(stripPrefix+"/api/torrents", s.handleTorrents)
	mux.HandleFunc(stripPrefix+"/api/logs", s.handleLogs)
	mux.HandleFunc(stripPrefix+"/api/delete", s.handleDelete)
	mux.HandleFunc(stripPrefix+"/api/candidates", s.handleCandidates)
	mux.HandleFunc(stripPrefix+"/api/history", s.handleHistory)
	mux.HandleFunc(stripPrefix+"/ws/logs", s.handleWebSocketLogs)

	// Static files and root
	mux.HandleFunc(stripPrefix+"/", s.handleRoot)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	s.logger.Infof("Starting web server on http://0.0.0.0:%d%s", s.port, stripPrefix)

	return s.srv.ListenAndServe()
}

// Stop stops the web server gracefully
func (s *Server) Stop() error {
	if s.srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

// handleStats returns current statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.cleaner.GetStats()
	if err != nil {
		s.logger.Errorf("Failed to get stats: %v", err)
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleTorrents returns list of torrents
func (s *Server) handleTorrents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	torrents, err := s.client.GetTorrents()
	if err != nil {
		s.logger.Errorf("Failed to get torrents: %v", err)
		http.Error(w, "Failed to get torrents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(torrents)
}

// handleLogs returns logs
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs := s.logger.GetLogs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// handleDelete handles torrent deletion
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing torrent id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid torrent id", http.StatusBadRequest)
		return
	}

	// Get torrent info first
	torrents, err := s.client.GetTorrents()
	if err != nil {
		http.Error(w, "Failed to get torrent info", http.StatusInternalServerError)
		return
	}

	var torrentName string
	var torrentTracker string
	var torrentSize int64
	for _, t := range torrents {
		if t.ID == id {
			torrentName = t.Name
			torrentTracker = t.NormalizedTracker
			torrentSize = t.TotalSize
			break
		}
	}

	// Delete torrent
	err = s.client.RemoveTorrent(id, true)
	if err != nil {
		s.logger.Errorf("Failed to delete torrent %d: %v", id, err)
		http.Error(w, "Failed to delete torrent", http.StatusInternalServerError)
		return
	}

	// Add to history
	s.cleaner.AddManualDeletion(id, torrentName, torrentTracker, torrentSize)

	s.logger.Infof("Manually deleted torrent: %s (ID: %d)", torrentName, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Torrent %d deleted successfully", id),
	})
}

// handleCandidates returns list of torrents that are candidates for deletion
func (s *Server) handleCandidates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	candidates, err := s.cleaner.GetCandidates()
	if err != nil {
		s.logger.Errorf("Failed to get candidates: %v", err)
		http.Error(w, "Failed to get candidates", http.StatusInternalServerError)
		return
	}

	// Return just the IDs of candidate torrents
	ids := make([]int, len(candidates))
	for i, t := range candidates {
		ids[i] = t.ID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

// handleHistory returns the deletion history
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history := s.cleaner.GetHistory()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// handleWebSocketLogs handles WebSocket connections for real-time logs
func (s *Server) handleWebSocketLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	s.clientMutex.Lock()
	s.logClients[conn] = true
	s.clientMutex.Unlock()

	s.logger.Debug("New WebSocket client connected")

	// Send existing logs
	logs := s.logger.GetLogs()
	for _, log := range logs {
		if err := conn.WriteJSON(log); err != nil {
			break
		}
	}

	// Keep connection alive and handle close
	defer func() {
		s.clientMutex.Lock()
		delete(s.logClients, conn)
		s.clientMutex.Unlock()
		conn.Close()
		s.logger.Debug("WebSocket client disconnected")
	}()

	// Read messages (just to detect disconnect)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// BroadcastLog broadcasts a log entry to all connected WebSocket clients
func (s *Server) BroadcastLog(log logger.LogEntry) {
	s.clientMutex.RLock()
	defer s.clientMutex.RUnlock()

	for client := range s.logClients {
		err := client.WriteJSON(log)
		if err != nil {
			s.logger.Debugf("Failed to send log to WebSocket client: %v", err)
		}
	}
}

// handleRoot serves the HTML interface
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Serve index.html for root
	if r.URL.Path == s.webRoot || r.URL.Path == strings.TrimSuffix(s.webRoot, "/") {
		s.serveHTML(w, r)
		return
	}

	// Serve static files
	staticPath := path.Join("web/static", strings.TrimPrefix(r.URL.Path, strings.TrimSuffix(s.webRoot, "/")))
	http.ServeFile(w, r, staticPath)
}

// serveHTML serves the main HTML interface
func (s *Server) serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Get webroot for use in HTML
	webRoot := strings.TrimSuffix(s.webRoot, "/")
	if webRoot == "" {
		webRoot = "/"
	}

	html := s.generateHTML(webRoot)
	w.Write([]byte(html))
}
