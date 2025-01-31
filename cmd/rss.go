package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"gofr.dev/pkg/gofr"
	"golang.org/x/exp/rand"
)

var (
	feed      *feeds.Feed
	feedMutex sync.Mutex
)

func initFeed() {
	feed = &feeds.Feed{
		Title:       "Docker Image Updates",
		Link:        &feeds.Link{Href: "http://localhost:8080/feed"},
		Description: "RSS feed for Docker image updates",
		Author:      &feeds.Author{Name: "Docker Updater", Email: "updater@localhost"},
		Created:     time.Now(),
	}
}

func generateRSSFeed(updates []ImageUpdate, c *gofr.Context) {
	feedMutex.Lock()
	defer feedMutex.Unlock()

	for _, update := range updates {
		if update.UpdateAvailable {
			c.Logf("update available for %s", update.ImageName)
			uniqueID := fmt.Sprintf("update-%s-%d-%d", update.ImageName, time.Now().UnixNano(), rand.Intn(10000))
			feed.Items = append(feed.Items, &feeds.Item{
				Title:       fmt.Sprintf("Update available for %s", update.ImageName),
				Link:        &feeds.Link{Href: "http://localhost:8080/feed"},
				Description: fmt.Sprintf("<b>Current Hash:</b> %s<br><b>Latest Hash:</b> %s<br><b>Architecture:</b> %s<br><b>Image Created:</b> %s<br>", update.CurrentHash, update.LatestHash, update.Architecture, update.ImageCreated),
				Created:     time.Now(),
				Id:          uniqueID,
			})
		}
	}

	if len(feed.Items) > 10 {
		feed.Items = feed.Items[len(feed.Items)-10:]
	}
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	feedMutex.Lock()
	defer feedMutex.Unlock()

	atom, err := feed.ToAtom()
	if err != nil {
		http.Error(w, "Error generating RSS feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml")
	w.Write([]byte(atom))
}
