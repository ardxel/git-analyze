package api

import (
	"fmt"
	"git-analyzer/pkg/analyzer"
	"git-analyzer/pkg/config"
	"git-analyzer/pkg/tasks"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v63/github"
)

var (
	githubClient = github.NewClient(nil).WithAuthToken(config.Vars.GithubPAT)
	githubRegexp = regexp.MustCompile(`^/([^/]+)/([^/]+)(?:/|$)`)
)

// GET /
func HandleGetForm(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", AnalyzeResultMap{
			RepoSizeLimit: config.Vars.MaxRepoSize, // MB
			ParallelMode:  config.Vars.UseFileWorkers,
			IsProd:        config.Vars.GoEnv == "production",
		})
	}
}

type JSONTaskInit struct {
	ID           string `json:"id"`
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
	Cache        bool   `json:"cache"`
	CacheKey     string `json:"cache_key"`
	Position     int    `json:"position"`
}

// POST /api
func HandleCreateTask(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		rOwnerRaw, _ := c.Get("repo_owner")
		rNameRaw, _ := c.Get("repo_name")
		rOwner := rOwnerRaw.(string)
		rName := rNameRaw.(string)

		repoSize, err := fetchRepoSize(rOwner, rName)

		if err != nil {
			c.Error(NewAnalyzeError(http.StatusNotFound, err.Error()))
			return
		}

		if repoSize > tasks.RepoTaskQueue.MaxRepoSize {
			msg := fmt.Sprintf("Repository size is too large <strong>%d MB</strong>", repoSize/(1024*1024))
			c.Error(NewAnalyzeError(http.StatusBadRequest, msg))
			return
		}

		repoTask := &tasks.RepoTask{
			Status: tasks.STATUS_INIT,
			Size:   repoSize,
			Owner:  rOwner,
			Name:   rName,
			Opts: &analyzer.Options{
				ExcludeFilePatterns: c.PostFormArray("exclude_file_patterns[]"),
				ExcludeDirPatterns:  c.PostFormArray("exclude_dir_patterns[]"),
			},
		}

		taskID := tasks.RepoTaskQueue.Add(repoTask)

		c.JSON(http.StatusAccepted, JSONTaskInit{
			ID:           taskID,
			Error:        false,
			ErrorMessage: "",
			Cache:        false,
			CacheKey:     "",
			Position:     -1,
		})
	}
}

const (
	ACTION_STATUS = "0"
	ACTION_RESULT = "1"
)

type JSONTaskStatus struct {
	TaskStatus       uint8  `json:"task_status"`
	TaskDone         bool   `json:"task_done"`
	TaskError        bool   `json:"task_error"`
	TaskErrorMessage string `json:"task_error_message"`
}

type AnalyzeResultMap struct {
	RepoSizeLimit int64                `redis:"repo_size_limit"`
	IsProd        bool                 `redis:"is_prod"`
	ParallelMode  bool                 `redis:"parallel_mode"`
	Languages     []*analyzer.Language `redis:"languages"`
	TotalLines    int32                `redis:"total_lines"`
	TotalFiles    int32                `redis:"total_files"`
	TotalBlank    int32                `redis:"total_blank"`
	TotalComments int32                `redis:"total_comments"`
	FetchSpeed    time.Duration        `redis:"fetch_speed"`
	AnalysisSpeed time.Duration        `redis:"analysis_speed"`
	Error         string               `redis:"error"`
}

// GET /api/task/:id/:action
// TODO: Add data about task queue
func HandleTask(s *Server) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		action := c.Param("action")

		task, ok := tasks.RepoTaskQueue.GetTask(id)

		if !ok {
			c.Error(NewTaskStatusError(nil, http.StatusNotFound, "Task not found"))
			return
		}

		switch action {
		case ACTION_STATUS:
			if task.Err != nil {
				c.Error(NewTaskStatusError(nil, http.StatusBadRequest, task.Err.Error()))
				return
			}

			c.JSON(http.StatusOK, JSONTaskStatus{
				TaskStatus:       task.Status,
				TaskDone:         task.Status == tasks.STATUS_DONE,
				TaskError:        false,
				TaskErrorMessage: "",
			})
		case ACTION_RESULT:
			if task.Status != tasks.STATUS_DONE {
				c.Error(NewTaskStatusError(task, http.StatusBadRequest, "Task not done"))
				return
			}

			tasks.RepoTaskQueue.DeleteTask(id)

			if task.Err != nil {
				c.Error(NewTaskStatusError(task, http.StatusBadRequest, task.Err.Error()))
				return
			}

			data := &AnalyzeResultMap{
				RepoSizeLimit: config.Vars.MaxRepoSize,
				ParallelMode:  config.Vars.UseFileWorkers,
				Languages:     task.Result.Languages,
				TotalLines:    task.Result.TotalLines,
				TotalFiles:    task.Result.TotalFiles,
				TotalBlank:    task.Result.TotalBlank,
				TotalComments: task.Result.TotalComments,
				FetchSpeed:    task.FetchSpeed,
				AnalysisSpeed: task.AnalysisSpeed,
			}

			keyForRedis, ok := RepoTaskResultKey(task.Owner, task.Name)

			if ok {
				s.Redis.SetCache(keyForRedis, data)
			}

			c.HTML(http.StatusOK, "table.html", data)
		default:
			c.Error(NewTaskStatusError(task, http.StatusNotFound, "Unknown action"))
		}

	}
}
