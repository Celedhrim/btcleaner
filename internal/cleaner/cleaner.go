package cleaner

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Celedhrim/btcleaner/internal/logger"
	"github.com/Celedhrim/btcleaner/internal/transmission"
	"github.com/Celedhrim/btcleaner/pkg/models"
)

const MaxHistorySize = 50

// DeletedTorrent represents a torrent that was deleted
type DeletedTorrent struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	SizeGB    float64   `json:"size_gb"`
	Tracker   string    `json:"tracker"`
	DeletedAt time.Time `json:"deleted_at"`
	Reason    string    `json:"reason"` // "auto" or "manual"
}

// Cleaner handles torrent cleanup logic
type Cleaner struct {
	client                *transmission.Client
	minFreeSpace          int64
	minTorrentsPerTracker int
	dryRun                bool
	logger                *logger.Logger
	history               []DeletedTorrent
	historyMutex          sync.RWMutex
}

// New creates a new Cleaner
func New(client *transmission.Client, minFreeSpace int64, minTorrentsPerTracker int, dryRun bool, log *logger.Logger) *Cleaner {
	return &Cleaner{
		client:                client,
		minFreeSpace:          minFreeSpace,
		minTorrentsPerTracker: minTorrentsPerTracker,
		dryRun:                dryRun,
		logger:                log,
		history:               make([]DeletedTorrent, 0, MaxHistorySize),
	}
}

// CleanupResult contains information about cleanup operation
type CleanupResult struct {
	InitialFreeSpace int64
	FinalFreeSpace   int64
	RemovedCount     int
	RemovedSize      int64
	RemovedTorrents  []models.Torrent
	NeedCleanup      bool
}

// Run executes the cleanup process
func (c *Cleaner) Run() (*CleanupResult, error) {
	result := &CleanupResult{}

	// Get current free space
	freeSpace, err := c.client.GetFreeSpace()
	if err != nil {
		return nil, fmt.Errorf("failed to get free space: %w", err)
	}

	result.InitialFreeSpace = freeSpace
	result.FinalFreeSpace = freeSpace

	c.logger.Debugf("Current free space: %.2f GB", float64(freeSpace)/(1024*1024*1024))
	c.logger.Debugf("Minimum required: %.2f GB", float64(c.minFreeSpace)/(1024*1024*1024))

	// Check if cleanup is needed
	if freeSpace >= c.minFreeSpace {
		c.logger.Debug("Free space is sufficient, no cleanup needed")
		result.NeedCleanup = false
		return result, nil
	}

	result.NeedCleanup = true
	spaceNeeded := c.minFreeSpace - freeSpace
	c.logger.Warnf("Need to free up %.2f GB", float64(spaceNeeded)/(1024*1024*1024))

	// Get all torrents
	torrents, err := c.client.GetTorrents()
	if err != nil {
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	c.logger.Infof("Found %d torrents", len(torrents))

	// Select torrents to remove
	toRemove, err := c.selectTorrentsToRemove(torrents, spaceNeeded)
	if err != nil {
		return nil, err
	}

	if len(toRemove) == 0 {
		c.logger.Warn("Cannot free enough space while respecting minimum torrents per tracker constraint")
		return result, nil
	}

	result.RemovedTorrents = toRemove
	result.RemovedCount = len(toRemove)

	// Calculate total size to be freed
	for _, t := range toRemove {
		result.RemovedSize += t.TotalSize
	}

	c.logger.Infof("Selected %d torrents to remove (will free %.2f GB)", 
		len(toRemove), float64(result.RemovedSize)/(1024*1024*1024))

	// Remove torrents
	if c.dryRun {
		c.logger.Info("DRY RUN: Would remove the following torrents:")
		for _, t := range toRemove {
			c.logger.Infof("  - [%s] %s (%.2f GB, added %s)",
				t.NormalizedTracker, t.Name, 
				float64(t.TotalSize)/(1024*1024*1024),
				t.AddedDate.Format("2006-01-02"))
		}
	} else {
		for _, t := range toRemove {
			c.logger.Infof("Removing torrent: [%s] %s (%.2f GB)", 
				t.NormalizedTracker, t.Name, float64(t.TotalSize)/(1024*1024*1024))
			
			if err := c.client.RemoveTorrent(t.ID, true); err != nil {
				c.logger.Errorf("Failed to remove torrent %s: %v", t.Name, err)
				continue
			}

			// Add to history
			c.addToHistory(t, "auto")
		}

		// Update final free space
		finalFreeSpace, err := c.client.GetFreeSpace()
		if err != nil {
			c.logger.Warnf("Failed to get final free space: %v", err)
		} else {
			result.FinalFreeSpace = finalFreeSpace
			c.logger.Infof("Final free space: %.2f GB", float64(finalFreeSpace)/(1024*1024*1024))
		}
	}

	return result, nil
}

// selectTorrentsToRemove selects torrents to remove based on age while respecting tracker minimums
func (c *Cleaner) selectTorrentsToRemove(torrents []models.Torrent, spaceNeeded int64) ([]models.Torrent, error) {
	// Group torrents by tracker
	trackerMap := make(map[string][]models.Torrent)
	for _, t := range torrents {
		trackerMap[t.NormalizedTracker] = append(trackerMap[t.NormalizedTracker], t)
	}

	// Sort torrents in each tracker by age (oldest first)
	for tracker := range trackerMap {
		sort.Sort(models.TorrentsByAge(trackerMap[tracker]))
	}

	c.logger.Debug("Torrent distribution by tracker:")
	for tracker, tList := range trackerMap {
		c.logger.Debugf("  %s: %d torrents", tracker, len(tList))
	}

	// Select torrents to remove
	var toRemove []models.Torrent
	var totalFreed int64

	// Create a copy of the tracker map to track remaining torrents
	remainingMap := make(map[string]int)
	for tracker, tList := range trackerMap {
		remainingMap[tracker] = len(tList)
	}

	// Sort all torrents by age
	allTorrents := make([]models.Torrent, len(torrents))
	copy(allTorrents, torrents)
	sort.Sort(models.TorrentsByAge(allTorrents))

	// Select oldest torrents that can be removed
	for _, t := range allTorrents {
		// Check if we've freed enough space
		if totalFreed >= spaceNeeded {
			break
		}

		// Check if we can remove this torrent (tracker has more than minimum)
		if remainingMap[t.NormalizedTracker] <= c.minTorrentsPerTracker {
			c.logger.Debugf("Cannot remove %s: tracker %s at minimum (%d torrents)", 
				t.Name, t.NormalizedTracker, remainingMap[t.NormalizedTracker])
			continue
		}

		// Add to removal list
		toRemove = append(toRemove, t)
		totalFreed += t.TotalSize
		remainingMap[t.NormalizedTracker]--

		c.logger.Debugf("Selected for removal: %s (tracker: %s, size: %.2f GB, remaining: %d)",
			t.Name, t.NormalizedTracker, float64(t.TotalSize)/(1024*1024*1024), 
			remainingMap[t.NormalizedTracker])
	}

	// Check if we could free enough space
	if totalFreed < spaceNeeded {
		c.logger.Warnf("Could only free %.2f GB out of %.2f GB needed", 
			float64(totalFreed)/(1024*1024*1024), 
			float64(spaceNeeded)/(1024*1024*1024))
	}

	return toRemove, nil
}

// GetCandidates returns torrents that would be deleted in a cleanup
func (c *Cleaner) GetCandidates() ([]*models.Torrent, error) {
	freeSpace, err := c.client.GetFreeSpace()
	if err != nil {
		return nil, fmt.Errorf("failed to get free space: %w", err)
	}

	// Check if cleanup is needed
	if freeSpace >= c.minFreeSpace {
		return []*models.Torrent{}, nil
	}

	torrents, err := c.client.GetTorrents()
	if err != nil {
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	// Calculate space needed to reach minimum
	spaceNeeded := c.minFreeSpace - freeSpace

	// Select torrents to remove (without actually removing them)
	candidates, err := c.selectTorrentsToRemove(torrents, spaceNeeded)
	if err != nil {
		return nil, fmt.Errorf("failed to select candidates: %w", err)
	}

	// Convert to pointers
	result := make([]*models.Torrent, len(candidates))
	for i := range candidates {
		result[i] = &candidates[i]
	}

	return result, nil
}

// GetStats returns current statistics
func (c *Cleaner) GetStats() (map[string]interface{}, error) {
	freeSpace, err := c.client.GetFreeSpace()
	if err != nil {
		return nil, err
	}

	torrents, err := c.client.GetTorrents()
	if err != nil {
		return nil, err
	}

	// Count torrents and space by tracker
	trackerCounts := make(map[string]int)
	trackerSpace := make(map[string]int64)
	var totalSpace int64 = 0
	
	for _, t := range torrents {
		trackerCounts[t.NormalizedTracker]++
		trackerSpace[t.NormalizedTracker] += t.TotalSize
		totalSpace += t.TotalSize
	}

	// Build tracker stats with count and space
	trackerStats := make([]map[string]interface{}, 0)
	for tracker, count := range trackerCounts {
		trackerStats = append(trackerStats, map[string]interface{}{
			"name":  tracker,
			"count": count,
			"size_gb": float64(trackerSpace[tracker]) / (1024 * 1024 * 1024),
		})
	}

	needsCleanup := freeSpace < c.minFreeSpace
	var spaceToRecover int64 = 0
	var candidatesCount int = 0

	// Get candidates if cleanup is needed
	if needsCleanup {
		candidates, err := c.GetCandidates()
		if err == nil {
			candidatesCount = len(candidates)
			for _, t := range candidates {
				spaceToRecover += t.TotalSize
			}
		}
	}

	stats := map[string]interface{}{
		"free_space_bytes":     freeSpace,
		"free_space_gb":        float64(freeSpace) / (1024 * 1024 * 1024),
		"min_free_space_gb":    float64(c.minFreeSpace) / (1024 * 1024 * 1024),
		"total_torrents":       len(torrents),
		"total_space_gb":       float64(totalSpace) / (1024 * 1024 * 1024),
		"tracker_stats":        trackerStats,
		"needs_cleanup":        needsCleanup,
		"candidates_count":     candidatesCount,
		"space_to_recover_gb":  float64(spaceToRecover) / (1024 * 1024 * 1024),
	}

	return stats, nil
}

// addToHistory adds a deleted torrent to the history
func (c *Cleaner) addToHistory(t models.Torrent, reason string) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()

	deleted := DeletedTorrent{
		ID:        t.ID,
		Name:      t.Name,
		Size:      t.TotalSize,
		SizeGB:    float64(t.TotalSize) / (1024 * 1024 * 1024),
		Tracker:   t.NormalizedTracker,
		DeletedAt: time.Now(),
		Reason:    reason,
	}

	// Add to front of history
	c.history = append([]DeletedTorrent{deleted}, c.history...)

	// Keep only last MaxHistorySize entries
	if len(c.history) > MaxHistorySize {
		c.history = c.history[:MaxHistorySize]
	}
}

// AddManualDeletion adds a manually deleted torrent to history
func (c *Cleaner) AddManualDeletion(id int, name, tracker string, size int64) {
	c.addToHistory(models.Torrent{
		ID:               id,
		Name:             name,
		NormalizedTracker: tracker,
		TotalSize:        size,
	}, "manual")
}

// GetHistory returns the deletion history
func (c *Cleaner) GetHistory() []DeletedTorrent {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()

	// Return a copy
	result := make([]DeletedTorrent, len(c.history))
	copy(result, c.history)
	return result
}
