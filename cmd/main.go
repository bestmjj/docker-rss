package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type ImageUpdate struct {
	gorm.Model
	ImageName       string
	CurrentHash     string
	LatestHash      string
	UpdateAvailable bool
	Architecture    string
	ImageCreated    string
}

type Image struct {
	Architecture string `json:"architecture"`
	Digest       string `json:"digest"`
}

type Repository struct {
	Creator int     `json:"creator"`
	ID      int     `json:"id"`
	Digest  string  `json:"digest"`
	Images  []Image `json:"images"`
}

func updates(containers []types.Container, ctx context.Context, cli *client.Client) []ImageUpdate {
	var updates []ImageUpdate

	for _, container := range containers {
		namespace, repository, tag := parseImageName(container.Image)
		imageName := fmt.Sprintf("%s/%s:%s", namespace, repository, tag)
		currentHash, arch, imageCreated := getCurrentHash(ctx, cli, imageName)

		latestHash, err := getLatestHash(namespace, repository, tag)
		fmt.Println("namespace: ", namespace)
		fmt.Println("repository: ", repository)
		fmt.Println("tag: ", tag)
		fmt.Println("arch: ", arch)

		fmt.Printf("current hash for %s is: %s", container.Image, currentHash)
		fmt.Printf("latest hash for %s is: %s", container.Image, latestHash)
		fmt.Println("tag: ", tag)

		if err != nil {
			log.Fatal(err)
		}

		updateAvailable := strings.SplitN(currentHash, ":", 2)[1] != strings.SplitN(latestHash, ":", 2)[1]

		updates = append(updates, ImageUpdate{
			ImageName:       imageName,
			CurrentHash:     currentHash,
			LatestHash:      latestHash,
			UpdateAvailable: updateAvailable,
			Architecture:    arch,
			ImageCreated:    imageCreated,
		})
	}

	return updates
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{})
	if err != nil {
		panic(err)
	}

	initFeed()
	http.HandleFunc("/feed", feedHandler)

	// go func() {
	// 	imageUpdate := updates(containers, ctx, cli)
	// 	generateRSSFeed(imageUpdate)

	// 	ticker := time.NewTicker(30 * time.Hour)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		imageUpdate = updates(containers, ctx, cli)
	// 		generateRSSFeed(imageUpdate)
	// 	}
	// }()

	// fmt.Println("Server is running on http://localhost:8083/feed")
	// log.Fatal(http.ListenAndServe("0.0.0.0:8083", nil))

	go func() {
		log.Fatal(http.ListenAndServe(":8083", nil))
	}()

	cronSchedule := os.Getenv("UPDATE_SCHEDULE")
	if cronSchedule == "" {
		log.Println("cron not found")
		cronSchedule = "0 */24 * * * *"
	}

	c := cron.New()

	c.AddFunc(cronSchedule, func() {
		imageUpdate := updates(containers, ctx, cli)
		generateRSSFeed(imageUpdate)
	})

	c.Start()
}
