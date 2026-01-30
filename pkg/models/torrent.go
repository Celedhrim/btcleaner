package models

import "time"

// Torrent represents a torrent with its metadata
type Torrent struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Hash         string    `json:"hash"`
	AddedDate    time.Time `json:"addedDate"`
	TotalSize    int64     `json:"totalSize"`
	Trackers     []string  `json:"trackers"`
	NormalizedTracker string `json:"normalizedTracker"`
	Status       int       `json:"status"`
	PercentDone  float64   `json:"percentDone"`
}

// TorrentsByAge implements sort.Interface for []Torrent based on AddedDate
type TorrentsByAge []Torrent

func (a TorrentsByAge) Len() int           { return len(a) }
func (a TorrentsByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TorrentsByAge) Less(i, j int) bool { return a[i].AddedDate.Before(a[j].AddedDate) }
