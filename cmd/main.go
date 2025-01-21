package main

import (
	"context"
	"fmt"
	"log"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ImageUpdate struct {
	ImageName       string
	CurrentHash     string
	LatestHash      string
	UpdateAvailable bool
}

type Image struct {
	Digest string `json:"digest"`
}

type Repository struct {
	Creator int     `json:"creator"`
	ID      int     `json:"id"`
	Images  []Image `json:"images"`
}

var repository, tag, namespace string

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
	var updates []ImageUpdate

	for _, container := range containers {
		namespace, repository, tag := parseImageName(container.Image)
		imageName := fmt.Sprintf("%s/%s:%s", namespace, repository, tag)
		currentHash := getCurrentHash(ctx, cli, imageName)

		latestHash, err := getLatestHash(namespace, repository, tag)
		if err != nil {
			log.Fatal(err)
		}
		updateAvailable := currentHash != latestHash
		updates = append(updates, ImageUpdate{
			ImageName:       imageName,
			CurrentHash:     currentHash,
			LatestHash:      latestHash,
			UpdateAvailable: updateAvailable,
		})
	}

	generateRSSFeed(updates)
}
