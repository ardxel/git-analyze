package api

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
)

func validateRepoURL(rawurl string) bool {
	parsedURL, err := url.Parse(rawurl)

	if err != nil {
		return false
	}

	if parsedURL.Host != "github.com" {
		return false
	}

	matches := githubRegexp.FindStringSubmatch(parsedURL.Path)

	if matches == nil || len(matches) < 3 {
		return false
	}

	return true
}

func extractMeta(rawurl string) (onwer string, repo string, err error) {
	parsedURL, err := url.Parse(rawurl)

	if err != nil {
		return "", "", fmt.Errorf("invalid url: %s", err)
	}

	if parsedURL.Host != "github.com" {
		if parsedURL.Host == "" {
			parsedURL.Host = "none"
		}
		return "", "", fmt.Errorf("Wrong host, expected github.com, got %s", parsedURL.Host)
	}

	matches := githubRegexp.FindStringSubmatch(parsedURL.Path)

	if matches == nil || len(matches) < 3 {
		return "", "", fmt.Errorf("Url must contain owner and repo: %s", rawurl)
	}

	owner, repoName := matches[1], matches[2]

	return owner, repoName, nil
}

func getOwner(ctx *gin.Context) string {
	value, ok := ctx.Get("repo_owner")

	if !ok {
		return ""
	}

	return value.(string)
}

func getName(ctx *gin.Context) string {
	value, ok := ctx.Get("repo_name")

	if !ok {
		return ""
	}

	return value.(string)
}
