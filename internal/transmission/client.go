package transmission

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Celedhrim/btcleaner/pkg/models"
)

// Client is a Transmission RPC client
type Client struct {
	url      string
	username string
	password string
	client   *http.Client
	sessionID string
}

// NewClient creates a new Transmission client
func NewClient(url, username, password string) *Client {
	return &Client{
		url:      url,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// RPCRequest represents a Transmission RPC request
type RPCRequest struct {
	Method    string                 `json:"method"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Tag       int                    `json:"tag,omitempty"`
}

// RPCResponse represents a Transmission RPC response
type RPCResponse struct {
	Result    string                 `json:"result"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Tag       int                    `json:"tag,omitempty"`
}

// doRequest performs an RPC request
func (c *Client) doRequest(req *RPCRequest) (*RPCResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.username != "" {
		httpReq.SetBasicAuth(c.username, c.password)
	}
	if c.sessionID != "" {
		httpReq.Header.Set("X-Transmission-Session-Id", c.sessionID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle 409 Conflict (session ID required)
	if resp.StatusCode == 409 {
		c.sessionID = resp.Header.Get("X-Transmission-Session-Id")
		return c.doRequest(req) // Retry with session ID
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Result != "success" {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Result)
	}

	return &rpcResp, nil
}

// GetSessionStats gets session statistics including free space
func (c *Client) GetSessionStats() (map[string]interface{}, error) {
	req := &RPCRequest{
		Method: "session-stats",
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return resp.Arguments, nil
}

// GetSession gets session info including download directory
func (c *Client) GetSession() (map[string]interface{}, error) {
	req := &RPCRequest{
		Method: "session-get",
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return resp.Arguments, nil
}

// GetFreeSpace returns free space in bytes for the download directory
func (c *Client) GetFreeSpace() (int64, error) {
	// First get the download directory
	session, err := c.GetSession()
	if err != nil {
		return 0, fmt.Errorf("failed to get session: %w", err)
	}

	downloadDir, ok := session["download-dir"].(string)
	if !ok {
		return 0, fmt.Errorf("download-dir not found in session")
	}

	// Get free space for the download directory
	req := &RPCRequest{
		Method: "free-space",
		Arguments: map[string]interface{}{
			"path": downloadDir,
		},
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return 0, err
	}

	// Response contains "size-bytes" or "path" and "size-bytes"
	if sizeBytes, ok := resp.Arguments["size-bytes"].(float64); ok {
		return int64(sizeBytes), nil
	}

	return 0, fmt.Errorf("size-bytes not found in response")
}

// GetTorrents returns all torrents with their metadata
func (c *Client) GetTorrents() ([]models.Torrent, error) {
	req := &RPCRequest{
		Method: "torrent-get",
		Arguments: map[string]interface{}{
			"fields": []string{
				"id",
				"name",
				"hashString",
				"addedDate",
				"totalSize",
				"trackers",
				"status",
				"percentDone",
			},
		},
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	torrentsData, ok := resp.Arguments["torrents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("torrents not found in response")
	}

	torrents := make([]models.Torrent, 0, len(torrentsData))
	for _, td := range torrentsData {
		torrentMap, ok := td.(map[string]interface{})
		if !ok {
			continue
		}

		torrent := models.Torrent{
			ID:          int(torrentMap["id"].(float64)),
			Name:        torrentMap["name"].(string),
			Hash:        torrentMap["hashString"].(string),
			AddedDate:   time.Unix(int64(torrentMap["addedDate"].(float64)), 0),
			TotalSize:   int64(torrentMap["totalSize"].(float64)),
			Status:      int(torrentMap["status"].(float64)),
			PercentDone: torrentMap["percentDone"].(float64),
		}

		// Extract tracker URLs
		if trackersData, ok := torrentMap["trackers"].([]interface{}); ok {
			for _, t := range trackersData {
				if trackerMap, ok := t.(map[string]interface{}); ok {
					if announce, ok := trackerMap["announce"].(string); ok {
						torrent.Trackers = append(torrent.Trackers, announce)
					}
				}
			}
		}

		// Normalize tracker
		torrent.NormalizedTracker = normalizeTracker(torrent.Trackers)

		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

// normalizeTracker normalizes tracker URLs to a common name
func normalizeTracker(trackers []string) string {
	if len(trackers) == 0 {
		return "unknown"
	}

	// If multiple trackers, likely a public tracker
	if len(trackers) > 1 {
		return "public-tracker"
	}

	// Extract domain from tracker URL
	tracker := trackers[0]
	
	// Remove protocol
	tracker = strings.TrimPrefix(tracker, "http://")
	tracker = strings.TrimPrefix(tracker, "https://")
	tracker = strings.TrimPrefix(tracker, "udp://")
	
	// Extract domain (before first /)
	parts := strings.Split(tracker, "/")
	if len(parts) > 0 {
		domain := parts[0]
		// Remove port
		domain = strings.Split(domain, ":")[0]
		return domain
	}

	return "unknown"
}

// RemoveTorrent removes a torrent and its data
func (c *Client) RemoveTorrent(id int, deleteData bool) error {
	req := &RPCRequest{
		Method: "torrent-remove",
		Arguments: map[string]interface{}{
			"ids":               []int{id},
			"delete-local-data": deleteData,
		},
	}

	_, err := c.doRequest(req)
	return err
}

// RemoveTorrents removes multiple torrents and their data
func (c *Client) RemoveTorrents(ids []int, deleteData bool) error {
	if len(ids) == 0 {
		return nil
	}

	req := &RPCRequest{
		Method: "torrent-remove",
		Arguments: map[string]interface{}{
			"ids":               ids,
			"delete-local-data": deleteData,
		},
	}

	_, err := c.doRequest(req)
	return err
}

// TestConnection tests the connection to Transmission
func (c *Client) TestConnection() error {
	_, err := c.GetSession()
	return err
}
