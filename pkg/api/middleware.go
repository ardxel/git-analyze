package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ValidateFormMV(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		repoURL := c.PostForm("repo_url")

		owner, name, err := extractMeta(repoURL)

		if err != nil {
			owner = c.PostForm("repo_owner")
			name = c.PostForm("repo_name")

			if owner == "" || name == "" {
				c.Error(NewAnalyzeError(http.StatusBadRequest, err.Error()))
				c.Abort()
				return
			}
		}

		if !RepoIsExists(owner, name) {
			c.Error(NewAnalyzeError(http.StatusBadRequest, "Repository does not exist"))
			c.Abort()
			return
		}

		c.Set("repo_owner", owner)
		c.Set("repo_name", name)
		c.Next()
	}
}

func RedisRepoTaskCacheMV(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		owner := getOwner(c)
		name := getName(c)

		key, ok := RepoTaskResultKey(owner, name)

		if !ok {
			c.Next()
			return
		}

		cValue, cached := s.Redis.GetCache(key)

		if cached {
			data := cValue.(map[string]interface{})
			data["FetchSpeed"] = time.Duration(data["FetchSpeed"].(int64))
			data["AnalysisSpeed"] = time.Duration(data["AnalysisSpeed"].(int64))

			c.Header("X-Cache", "HIT")
			c.HTML(http.StatusOK, "table.html", data)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RedisRateLimitMV(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ip:%s", c.ClientIP())

		res := s.Redis.RateLimitAllow(key)
		if res.Allowed == 0 {
			retryAfterSeconds := int(res.ResetAfter / time.Second)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"id":            "",
				"error":         true,
				"error_message": fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", retryAfterSeconds),
				"cache":         false,
				"cache_key":     "",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CSP() func(*gin.Context) {
	cspPairs := (map[string]string{
		"default-src":     "'self'",
		"script-src":      "'self' https://cdn.example.com https://cdnjs.cloudflare.com",
		"style-src":       "'self' https://fonts.googleapis.com",
		"img-src":         "'self' data: https://img.shields.io https://github.githubassets.com",
		"font-src":        "'self' https://fonts.gstatic.com",
		"connect-src":     "'self'",
		"frame-ancestors": "'self'",
		"object-src":      "none",
		"base-uri":        "'self'",
		"form-action":     "'self'",
	})

	cspValue := ""

	for key, value := range cspPairs {
		cspValue += fmt.Sprintf("%s %s; ", key, value)
	}

	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", cspValue)
	}
}
