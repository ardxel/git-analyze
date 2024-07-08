package main

import (
	gitAnalyzer "git-analyzer/src/core"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	RepoURL             string   `json:"repo_url" binding:"required`
	ExcludeFilePatterns []string `json:"exclude_file_patterns"`
	ExcludeDirPatterns  []string `json:"exclude_dir_patterns"`
}

func main() {
	// defaultOptions := &gitAnalyzer.RepoValidateOptions{
	// 	ExcludeFilePatterns: []string{
	// 		"package*.json",
	// 		".git",
	// 	},
	// }

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		q := &QueryParams{}

		if err := c.ShouldBindJSON(q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		analyzeOptions := &gitAnalyzer.RepoValidateOptions{
			ExcludeFilePatterns: q.ExcludeFilePatterns,
			ExcludeDirPatterns:  q.ExcludeDirPatterns,
		}

		result, err := gitAnalyzer.AnalyzeRepository(q.RepoURL, analyzeOptions)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result})
	})

	r.Run("localhost:8080")
}
