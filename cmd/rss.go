package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/feeds"
)

func generateRSSFeed(updates []ImageUpdate) {
	feed := &feeds.Feed{
		Title:       "Docker Image Updates",
		Link:        &feeds.Link{Href: "http://localhost"},
		Description: "RSS feed for Docker image updates",
		Author:      &feeds.Author{Name: "Docker Updater", Email: "updater@localhost"},
		Created:     time.Now(),
	}

	feed.Items = []*feeds.Item{}

	for _, update := range updates {
		if update.UpdateAvailable {
			feed.Items = append(feed.Items, &feeds.Item{
				Title:       fmt.Sprintf("Update available for %s", update.ImageName),
				Link:        &feeds.Link{Href: "http://localhost"},
				Description: fmt.Sprintf("Current Hash: %s, Latest Hash: %s", update.CurrentHash, update.LatestHash),
				Created:     time.Now(),
			})
		}
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Fatalf("Error generating RSS feed: %v", err)
	}

	file, err := os.Create("docker_updates.rss")
	if err != nil {
		log.Fatalf("Error creating RSS file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(rss)
	if err != nil {
		log.Fatalf("Error writing RSS feed to file: %v", err)
	}

	fmt.Println("RSS feed generated successfully.")
}
