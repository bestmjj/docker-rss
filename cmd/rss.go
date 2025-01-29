package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"golang.org/x/exp/rand"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	feed      *feeds.Feed
	feedMutex sync.Mutex
	db        *gorm.DB
)

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("docker_updates.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = db.AutoMigrate(&ImageUpdate{})
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}
}

func initFeed() {
	feed = &feeds.Feed{
		Title:       "Docker Image Updates",
		Link:        &feeds.Link{Href: "http://localhost:8080/feed"},
		Description: "RSS feed for Docker image updates",
		Author:      &feeds.Author{Name: "Docker Updater", Email: "updater@localhost"},
		Created:     time.Now(),
	}
}

func loadFeedFromDB() {
	var updates []ImageUpdate
	result := db.Find(&updates)
	if result.Error != nil {
		log.Fatalf("Error querying database: %v", result.Error)
	}

	for _, update := range updates {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       fmt.Sprintf("Update available for %s", update.ImageName),
			Link:        &feeds.Link{Href: "http://localhost:8080/feed"},
			Description: fmt.Sprintf("Current Hash: %s, Latest Hash: %s", update.CurrentHash, update.LatestHash),
			Created:     update.CreatedAt,
			Id:          fmt.Sprintf("update-%s-%d", update.ImageName, update.CreatedAt.UnixNano()),
		})
	}
}

func saveUpdateToDB(update ImageUpdate) {
	result := db.Create(&update)
	if result.Error != nil {
		log.Fatalf("Error inserting update: %v", result.Error)
	}
}

func generateRSSFeed(updates []ImageUpdate) {
	feedMutex.Lock()
	defer feedMutex.Unlock()

	for _, update := range updates {
		if update.UpdateAvailable {
			uniqueID := fmt.Sprintf("update-%s-%d-%d", update.ImageName, time.Now().UnixNano(), rand.Intn(10000))
			feed.Items = append(feed.Items, &feeds.Item{
				Title:       fmt.Sprintf("Update available for %s", update.ImageName),
				Link:        &feeds.Link{Href: "http://localhost:8080/feed"},
				Description: fmt.Sprintf("Current Hash: %s, Latest Hash: %s", update.CurrentHash, update.LatestHash),
				Created:     time.Now(),
				Id:          uniqueID,
			})
			saveUpdateToDB(update)
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
