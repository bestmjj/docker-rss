package main

import (
        "context"
        "fmt"
        "log"
        "net/http"
        "os"
        "strings"
        "sync"

        "github.com/docker/docker/api/types"
        containertypes "github.com/docker/docker/api/types/container"
        "github.com/docker/docker/client"
        "github.com/robfig/cron"
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

var offlineImages = make(map[string]bool)

func checkOffline(container types.Container, namespace, repository, tag string) bool {
        statusCode, _ := pingDockerhub(namespace, repository, tag)
        if statusCode == http.StatusNotFound {
                offlineImages[container.Image] = true
                log.Printf("image %s is not available on dockerhub. likely local image.", container.Image)
                return true
        }
        return false
}

func updates(containers []types.Container, ctx context.Context, cli *client.Client) []ImageUpdate {
        var updates []ImageUpdate
        var wg sync.WaitGroup
        results := make(chan ImageUpdate, len(containers))
        for _, container := range containers {
                wg.Add(1)
                go func(container types.Container) {
                        defer wg.Done()
                        namespace, repository, tag := parseImageName(container.Image)
                        // check for offline images
                        for range offlineImages {
                                if offlineImages[container.Image] {
                                        return
                                }
                        }
                        if checkOffline(container, namespace, repository, tag) {
                                return
                        }
                        imageName := fmt.Sprintf("%s/%s:%s", namespace, repository, tag)
                        currentHash, arch, imageCreated := getCurrentHash(ctx, cli, imageName)
                        log.Printf("checking updates for %s", imageName)
                        latestHash, err := getLatestHash(namespace, repository, tag, arch)
                        log.Printf("namespace: %s, repository: %s, tag: %s, arch: %s", namespace, repository, tag, arch)
                        log.Printf("current hash for %s is: %s", container.Image, currentHash)
                        log.Printf("latest hash for %s is: %s", container.Image, latestHash)
                        if err != nil {
                                log.Printf("Error getting latest hash for %s: %v", imageName, err)
                                results <- ImageUpdate{
                                        ImageName:       imageName,
                                        CurrentHash:     currentHash,
                                        LatestHash:      latestHash,
                                        UpdateAvailable: false,
                                        Architecture:    arch,
                                        ImageCreated:    imageCreated,
                                }
                                return
                        }

                        // 修复：安全地比较哈希值
                        updateAvailable := false
                        if currentHash != "" && latestHash != "" {
                                // 提取当前哈希的核心部分
                                currentDigest := currentHash
                                if idx := strings.Index(currentHash, "@"); idx != -1 {
                                        currentDigest = currentHash[idx+1:] // 取 @ 之后的部分
                                }
                                // 确保当前哈希和最新哈希格式一致
                                if !strings.HasPrefix(currentDigest, "sha256:") && strings.HasPrefix(latestHash, "sha256:") {
                                        currentDigest = "sha256:" + currentDigest
                                }

                                updateAvailable = currentDigest != latestHash
                                if updateAvailable {
                                        log.Printf("Update available for %s: current=%s, latest=%s", imageName, currentDigest, latestHash)
                                } else {
                                        log.Printf("No update needed for %s: current=%s, latest=%s", imageName, currentDigest, latestHash)
                                }
                        } else if currentHash == "" {
                                log.Printf("Skipping update for %s: current hash is empty", imageName)
                        } else if latestHash == "" {
                                log.Printf("Skipping update for %s: latest hash is empty", imageName)
                        }

                        results <- ImageUpdate{
                                ImageName:       imageName,
                                CurrentHash:     currentHash,
                                LatestHash:      latestHash,
                                UpdateAvailable: updateAvailable,
                                Architecture:    arch,
                                ImageCreated:    imageCreated,
                        }
                }(container)
        }
        go func() {
                wg.Wait()
                close(results)
        }()
        for result := range results {
                updates = append(updates, result)
        }
        return updates
}

func cronJob(containers []types.Container, ctx context.Context, cli *client.Client) {
        imageUpdate := updates(containers, ctx, cli)
        updateImages(imageUpdate)
        //generateRSSFeed(imageUpdate)
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
                log.Fatal("panic containers: ", err)
        }

        initFeed()
        http.HandleFunc("/feed", feedHandler)

        go func() {
                log.Fatal(http.ListenAndServe("0.0.0.0:8083", nil))
        }()

        log.Println("docker-rss server started at 0.0.0.0:8083...")
        log.Printf("cronjob expression specified: %s", os.Getenv("UPDATE_SCHEDULE"))

        c := cron.New()

        c.AddFunc(os.Getenv("UPDATE_SCHEDULE"), func() {
                cronJob(containers, ctx, cli)
        })

        c.Start()
        select {}
}
