package nyaa

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"link-anime/internal/models"
)

const (
	nyaaBaseURL = "https://nyaa.si"
	rssPath     = "/?page=rss"
)

// rssItem represents a single item in the Nyaa RSS feed.
type rssItem struct {
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	GUID     string `xml:"guid"`
	Seeders  string `xml:"https://nyaa.si/xmlns/nyaa seeders"`
	Leechers string `xml:"https://nyaa.si/xmlns/nyaa leechers"`
	Size     string `xml:"https://nyaa.si/xmlns/nyaa size"`
}

type rssChannel struct {
	Items []rssItem `xml:"channel>item"`
}

var magnetRe = regexp.MustCompile(`magnet:\?xt=urn:btih:[a-fA-F0-9]{40}`)

// Search queries Nyaa's RSS feed for anime torrents.
func Search(query string, filter string) ([]models.NyaaResult, error) {
	params := url.Values{
		"page": {"rss"},
		"q":    {query},
		"c":    {"1_2"}, // Anime - English-translated
		"f":    {"0"},   // No filter
	}

	if filter == "trusted" {
		params.Set("f", "2")
	} else if filter == "noremakes" {
		params.Set("f", "1")
	}

	reqURL := nyaaBaseURL + "/?" + params.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("nyaa search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("nyaa returned status %d", resp.StatusCode)
	}

	var rss rssChannel
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("nyaa parse: %w", err)
	}

	var results []models.NyaaResult
	for _, item := range rss.Items {
		seeders, _ := strconv.Atoi(item.Seeders)
		leechers, _ := strconv.Atoi(item.Leechers)

		// Extract magnet from link or GUID (Nyaa provides torrent links, not magnets in RSS)
		// We'll provide the download link; the handler can fetch magnet if needed
		magnet := item.Link
		if strings.HasPrefix(item.GUID, "https://nyaa.si/view/") {
			// We'll construct a magnet-fetch URL; actual magnet requires scraping the page
			// For now, use the torrent download link
			magnet = item.Link
		}

		results = append(results, models.NyaaResult{
			Title:    item.Title,
			Magnet:   magnet,
			Size:     item.Size,
			Seeders:  seeders,
			Leechers: leechers,
		})
	}

	return results, nil
}

// SearchWithMagnets is like Search but also fetches magnet links from each result page.
// This is slower but provides actual magnet URIs.
func SearchWithMagnets(query string, filter string, limit int) ([]models.NyaaResult, error) {
	results, err := Search(query, filter)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for i := range results {
		// The Link field from RSS is the .torrent download URL
		// We need to get the page URL and scrape the magnet from it
		pageURL := results[i].Magnet
		if !strings.HasPrefix(pageURL, "https://nyaa.si/view/") {
			// Try to extract the view URL from the download link
			// Download links look like: https://nyaa.si/download/1234567.torrent
			pageURL = strings.Replace(pageURL, "/download/", "/view/", 1)
			pageURL = strings.TrimSuffix(pageURL, ".torrent")
		}

		resp, err := client.Get(pageURL)
		if err != nil {
			continue
		}

		body := make([]byte, 64*1024) // Read up to 64KB
		n, _ := resp.Body.Read(body)
		resp.Body.Close()

		if m := magnetRe.Find(body[:n]); m != nil {
			results[i].Magnet = string(m)
		}
	}

	return results, nil
}

// FetchRSSFeed fetches a custom Nyaa RSS URL and returns parsed results.
func FetchRSSFeed(rssURL string) ([]models.NyaaResult, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(rssURL)
	if err != nil {
		return nil, fmt.Errorf("fetch rss: %w", err)
	}
	defer resp.Body.Close()

	var rss rssChannel
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}

	var results []models.NyaaResult
	for _, item := range rss.Items {
		seeders, _ := strconv.Atoi(item.Seeders)
		leechers, _ := strconv.Atoi(item.Leechers)

		results = append(results, models.NyaaResult{
			Title:    item.Title,
			Magnet:   item.Link,
			Size:     item.Size,
			Seeders:  seeders,
			Leechers: leechers,
		})
	}

	return results, nil
}
