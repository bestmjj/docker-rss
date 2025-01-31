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
	"gofr.dev/pkg/gofr"
)

type ImageUpdate struct {
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

func updates(containers []types.Container, ctx context.Context, cli *client.Client, c *gofr.Context) []ImageUpdate {
	var updates []ImageUpdate

	for _, container := range containers {
		namespace, repository, tag := parseImageName(container.Image, c)

		// a very unreliable hack to exclude an image which is not on dockerhub
		if strings.Contains(container.Image, "docker-rss") {
			continue
		}
		imageName := fmt.Sprintf("%s/%s:%s", namespace, repository, tag)
		currentHash, arch, imageCreated := getCurrentHash(ctx, cli, imageName)

		c.Logf("checking updates for %s", imageName)
		latestHash, err := getLatestHash(namespace, repository, tag, c)
		c.Logf("namespace: %s, repository: %s, tag: %s, arch: %s", namespace, repository, tag, arch)

		c.Logf("current hash for %s is: %s", container.Image, currentHash)
		c.Logf("latest hash for %s is: %s", container.Image, latestHash)

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

	go func() {
		log.Fatal(http.ListenAndServe("0.0.0.0:8083", nil))
	}()

	app := gofr.New()
	app.Logger().Log("docker-rss updater server started at 0.0.0.0:8083...")

	app.AddCronJob(os.Getenv("UPDATE_SCHEDULE"), "cron-dockerss", func(c *gofr.Context) {
		imageUpdate := updates(containers, ctx, cli, c)
		generateRSSFeed(imageUpdate, c)
	})

	app.Run()
}
