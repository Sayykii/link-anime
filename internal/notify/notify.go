package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Notifier sends notifications to configured endpoints.
type Notifier struct {
	URL string
}

// New creates a new Notifier. Returns nil if URL is empty.
func New(url string) *Notifier {
	if url == "" {
		return nil
	}
	return &Notifier{URL: url}
}

// Field is a key-value pair for notification extras.
type Field struct {
	Name  string
	Value string
}

// Send dispatches a notification. Auto-detects the service from the URL.
func (n *Notifier) Send(title, message string, fields []Field, color string) {
	if n == nil || n.URL == "" {
		return
	}

	if strings.Contains(n.URL, "ntfy") {
		n.sendNtfy(title, message, fields)
	} else if strings.Contains(n.URL, "discord") {
		n.sendDiscord(title, message, fields, color)
	} else {
		n.sendGeneric(title, message)
	}
}

func (n *Notifier) sendNtfy(title, message string, fields []Field) {
	body := message
	for _, f := range fields {
		body += "\n" + f.Name + ": " + f.Value
	}

	req, err := http.NewRequest("POST", n.URL, strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Title", title)

	client := &http.Client{Timeout: 10 * time.Second}
	client.Do(req) //nolint:errcheck
}

func (n *Notifier) sendDiscord(title, message string, fields []Field, color string) {
	embedColor := 3066993 // green
	switch color {
	case "blue":
		embedColor = 3447003
	case "red":
		embedColor = 15158332
	case "green":
		embedColor = 3066993
	}

	type embedField struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline"`
	}

	type embed struct {
		Title       string       `json:"title"`
		Description string       `json:"description"`
		Color       int          `json:"color"`
		Fields      []embedField `json:"fields,omitempty"`
		Footer      struct {
			Text string `json:"text"`
		} `json:"footer"`
		Timestamp string `json:"timestamp"`
	}

	var ef []embedField
	for _, f := range fields {
		ef = append(ef, embedField{Name: f.Name, Value: f.Value, Inline: true})
	}

	e := embed{
		Title:       title,
		Description: message,
		Color:       embedColor,
		Fields:      ef,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	e.Footer.Text = "link-anime"

	payload := map[string]interface{}{
		"embeds": []embed{e},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", n.URL, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	client.Do(req) //nolint:errcheck
}

func (n *Notifier) sendGeneric(title, message string) {
	payload := map[string]string{
		"title":   title,
		"message": message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", n.URL, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	client.Do(req) //nolint:errcheck
}

// FormatSize converts bytes to a human-readable string.
func FormatSize(bytes int64) string {
	const (
		GB = 1073741824
		MB = 1048576
		KB = 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
