package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/docker/docker/client"
	"gofr.dev/pkg/gofr"
)

func getCurrentHash(ctx context.Context, cli *client.Client, imageName string) (string, string, string) {
	image, _, err := cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		log.Printf("Error inspecting image %s: %v", imageName, err)
		return "", "", ""
	}
	return image.RepoDigests[0], image.Architecture, image.Created
}

func getLatestHash(namespace, repository, tag string, c *gofr.Context) (string, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/%s/repositories/%s/tags/%s", namespace, repository, tag)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	c.Logf("checking on dockerhub: %s", url)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var repo Repository
	err = json.Unmarshal(body, &repo)
	if err != nil {
		return "", err
	}

	// for _, image := range repo.Images {
	// 	if image.Architecture == architecture {
	// 		return image.Digest, nil
	// 	}
	// }

	return repo.Digest, nil

	// return "", fmt.Errorf("no images found for architecture: %s", architecture)
}
