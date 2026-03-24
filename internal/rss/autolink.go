package rss

import (
	"fmt"
	"log"

	"link-anime/internal/linker"
	"link-anime/internal/models"
	"link-anime/internal/notify"
	"link-anime/internal/ws"
)

// AutoLinker handles automatic linking of completed RSS-matched downloads.
type AutoLinker struct {
	Hub         *ws.Hub
	Notifier    func() *notify.Notifier
	DownloadDir func() string
	MediaDir    func() string
	MoviesDir   func() string
	ShokoScan   func() // trigger shoko to import new files
}

// HandleCompletion is called when the download monitor detects a completed torrent.
// It checks if the torrent came from an RSS match and auto-links it.
func (a *AutoLinker) HandleCompletion(t models.TorrentStatus) {
	match, rule := FindMatchByTorrentName(t.Name)
	if match == nil {
		return // not from RSS
	}

	if !rule.AutoLink {
		log.Printf("[autolink] skipping %q (auto-link disabled for rule %q)", t.Name, rule.Name)
		return
	}

	log.Printf("[autolink] auto-linking %q → %s S%02d", t.Name, rule.ShowName, rule.Season)

	req := models.LinkRequest{
		Source: t.Name,
		Type:   rule.MediaType,
		Name:   rule.ShowName,
		Season: rule.Season,
	}

	result, err := linker.Link(req, a.DownloadDir(), a.MediaDir(), a.MoviesDir(), a.Hub)
	if err != nil {
		log.Printf("[autolink] failed to link %q: %v", t.Name, err)
		UpdateMatchStatus(match.ID, "failed")

		a.Hub.Broadcast(models.WSMessage{
			Type: "autolink_failed",
			Data: map[string]interface{}{
				"title":  t.Name,
				"rule":   rule.Name,
				"error":  err.Error(),
			},
		})
		return
	}

	if result.Linked == 0 {
		log.Printf("[autolink] no files linked for %q (skipped=%d)", t.Name, result.Skipped)
		return
	}

	UpdateMatchStatus(match.ID, "linked")
	log.Printf("[autolink] linked %d files for %q → %s", result.Linked, t.Name, result.DestDir)

	// Broadcast success
	a.Hub.Broadcast(models.WSMessage{
		Type: "autolink_complete",
		Data: map[string]interface{}{
			"title":   t.Name,
			"rule":    rule.Name,
			"linked":  result.Linked,
			"destDir": result.DestDir,
		},
	})

	// Send notification
	if n := a.Notifier(); n != nil {
		title := "Auto-linked: " + rule.ShowName
		msg := ""
		if rule.MediaType == "series" {
			msg = fmt.Sprintf("Linked to Season %d", rule.Season)
		} else {
			msg = "Linked movie"
		}
		n.Send(title, msg, []notify.Field{
			{Name: "Files", Value: fmt.Sprintf("%d", result.Linked)},
			{Name: "Size", Value: notify.FormatSize(result.Size)},
			{Name: "Source", Value: t.Name},
		}, "green")
	}

	// Trigger Shoko to import new files
	if a.ShokoScan != nil {
		go a.ShokoScan()
	}
}
