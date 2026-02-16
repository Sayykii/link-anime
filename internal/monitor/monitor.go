package monitor

import (
	"log"
	"sync"
	"time"

	"link-anime/internal/models"
	"link-anime/internal/notify"
	"link-anime/internal/qbit"
	"link-anime/internal/ws"
)

// DownloadMonitor polls qBittorrent for torrent progress and broadcasts
// updates via WebSocket. It also sends notifications when downloads complete.
type DownloadMonitor struct {
	hub        *ws.Hub
	qbitGetter func() *qbit.Client
	notifier   func() *notify.Notifier
	category   func() string
	interval   time.Duration

	// Track previous torrent states to detect completions
	prevStates map[string]float64
	mu         sync.Mutex

	stopCh chan struct{}
	done   chan struct{}
}

// NewDownloadMonitor creates a new monitor instance.
// qbitGetter and notifier are functions so they pick up reinitClients() changes.
// category is a function that returns the current qBit category from settings.
func NewDownloadMonitor(
	hub *ws.Hub,
	qbitGetter func() *qbit.Client,
	notifier func() *notify.Notifier,
	category func() string,
	interval time.Duration,
) *DownloadMonitor {
	return &DownloadMonitor{
		hub:        hub,
		qbitGetter: qbitGetter,
		notifier:   notifier,
		category:   category,
		interval:   interval,
		prevStates: make(map[string]float64),
		stopCh:     make(chan struct{}),
		done:       make(chan struct{}),
	}
}

// Start begins the monitoring loop in a goroutine.
func (m *DownloadMonitor) Start() {
	go m.run()
}

// Stop signals the monitor to shut down and waits for it.
func (m *DownloadMonitor) Stop() {
	close(m.stopCh)
	<-m.done
}

func (m *DownloadMonitor) run() {
	defer close(m.done)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Do an initial poll right away
	m.poll()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.poll()
		}
	}
}

func (m *DownloadMonitor) poll() {
	client := m.qbitGetter()
	if client == nil || !client.IsConfigured() {
		return
	}

	cat := m.category()
	torrents, err := client.ListTorrents(cat)
	if err != nil {
		log.Printf("[monitor] failed to poll qBit: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var completed []models.TorrentStatus

	for _, t := range torrents {
		prevProgress, tracked := m.prevStates[t.Hash]

		// Detect newly completed torrents: was being tracked AND progress went from <1 to 1
		if tracked && prevProgress < 1.0 && t.Progress >= 1.0 {
			completed = append(completed, t)
		}

		m.prevStates[t.Hash] = t.Progress
	}

	// Clean up old hashes no longer in the torrent list
	currentHashes := make(map[string]bool, len(torrents))
	for _, t := range torrents {
		currentHashes[t.Hash] = true
	}
	for hash := range m.prevStates {
		if !currentHashes[hash] {
			delete(m.prevStates, hash)
		}
	}

	// Broadcast torrent progress to all WebSocket clients
	m.hub.Broadcast(models.WSMessage{
		Type: "torrent_progress",
		Data: models.TorrentProgress{
			Torrents:  torrents,
			Completed: completed,
		},
	})

	// Send notifications for newly completed downloads
	if len(completed) > 0 {
		n := m.notifier()
		for _, t := range completed {
			log.Printf("[monitor] download complete: %s", t.Name)

			// Broadcast a specific completion event
			m.hub.Broadcast(models.WSMessage{
				Type: "download_complete",
				Data: t,
			})

			// Send external notification
			if n != nil {
				n.Send(
					"Download Complete",
					t.Name,
					[]notify.Field{
						{Name: "Size", Value: notify.FormatSize(t.Size)},
					},
					"blue",
				)
			}
		}
	}
}
