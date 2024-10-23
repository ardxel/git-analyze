package api

import (
	"context"
	"fmt"
	"git-analyzer/pkg/config"
	"log"

	"github.com/google/go-github/v63/github"
)

func RepoIsExists(owner, name string) bool {
	ctx := context.Background()

	_, _, err := githubClient.Repositories.Get(ctx, owner, name)

	if _, ok := getRateLimitError(err); ok {
		return true
	}

	return err == nil
}

// returns repo size in bytes
func fetchRepoSize(owner, name string) (int64, error) {

	ctx := context.Background()
	repo, res, err := githubClient.Repositories.Get(ctx, owner, name)

	if errRateLimit, ok := getRateLimitError(err); ok {
		return 0, errRateLimit
	}

	if res.StatusCode == 404 {
		return 0, fmt.Errorf("Repository not found")
	}

	repoSize := int64(*repo.Size) * 1024 // bytes

	if config.Vars.Debug {
		log.Printf("Repo: %v, Size: %d MB\n", name, repoSize/1024/1024)
	}

	return repoSize, nil
}

func getRateLimitError(err error) (error, bool) {
	if err != nil {
		if rle, rateLimitOk := err.(*github.RateLimitError); rateLimitOk {
			return fmt.Errorf("GitHub API rate limit exceeded. Try again in %s", rle.Rate.Reset), true
		}
		if arle, abuseRateLimitOk := err.(*github.AbuseRateLimitError); abuseRateLimitOk {
			return fmt.Errorf("GitHub API rate limit exceeded. Try again in %s", arle.RetryAfter), true
		}
	}
	return nil, false
}
