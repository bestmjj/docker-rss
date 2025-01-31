package main

import (
	"strings"

	"gofr.dev/pkg/gofr"
)

func parseImageName(imageName string, c *gofr.Context) (string, string, string) {
	namespace := "library"
	repository := ""
	tag := "latest"

	if !strings.Contains(imageName, "/") && !strings.Contains(imageName, ":") {
		c.Logf("image [%s] does not have ns and tag", imageName)
		repository = imageName
	} else if !strings.Contains(imageName, "/") {
		c.Logf("image [%s] does not have ns", imageName)
		parts := strings.SplitN(imageName, ":", 2)
		repository = parts[0]
		if len(parts) == 2 {
			tag = parts[1]
		}
	} else {
		parts := strings.SplitN(imageName, ":", 2)
		fullValue := parts[0]
		if len(parts) == 2 {
			tag = parts[1]
		}

		nameParts := strings.SplitN(fullValue, "/", 2)
		namespace = nameParts[0]
		repository = nameParts[1]
	}

	return namespace, repository, tag
}
