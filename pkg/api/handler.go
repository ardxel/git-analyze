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
	githubClient = github.NewClient(nil).WithAuthToken(config.Vars.GithubApiPat)
	githubRegexp = regexp.MustCompile(`^/([^/]+)/([^/]+)(?:/|$)`)
)

type ResponseData struct {
	RepoSizeLimit   int64                `redis:"repo_size_limit" json:"repo_size_limit"`
	IsProd          bool                 `redis:"is_prod" json:"is_prod"`
	ParallelMode    bool                 `redis:"parallel_mode" json:"parallel_mode"`
	Languages       []*analyzer.Language `redis:"languages" json:"languages"`
	TotalLines      int32                `redis:"total_lines" json:"total_lines"`
	TotalFiles      int32                `redis:"total_files" json:"total_files"`
	TotalBlank      int32                `redis:"total_blank" json:"total_blank"`
	TotalComments   int32                `redis:"total_comments" json:"total_comments"`
	FetchSpeed      time.Duration        `redis:"fetch_speed" json:"fetch_speed"`
	AnalysisSpeed   time.Duration        `redis:"analysis_speed" json:"analysis_speed"`
	FetchSpeedStr   string               `redis:"fetch_speed_str" json:"fetch_speed_str"`
	AnalysisSpeeStr string               `redis:"analysis_speed_str" json:"analysis_speed_str"`
	Error           string               `redis:"error" json:"error"`
}

// GET /
func HandleGetForm(s *Server) func(*gin.Context) {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", ResponseData{
			RepoSizeLimit: config.Vars.MaxRepoSize, // MB
			ParallelMode:  config.Vars.UseFileWorkers,
			IsProd:        config.Vars.GoEnv == "production",
		})
	}
}

type TaskInit struct {
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

		c.JSON(http.StatusAccepted, TaskInit{
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

type TaskInfo struct {
	Status       uint8         `json:"task_status"`
	Done         bool          `json:"task_done"`
	Error        bool          `json:"task_error"`
	ErrorMessage string        `json:"task_error_message"`
	Result       *ResponseData `json:"task_result"`
}

// GET /api/task/:id/:action
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

			c.JSON(http.StatusOK, TaskInfo{
				Status:       task.Status,
				Done:         task.Status == tasks.STATUS_DONE,
				Error:        false,
				ErrorMessage: "",
				Result:       nil,
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

			data := &ResponseData{
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

			switch c.GetHeader("Accept") {
			case "application/json":

				for _, lang := range data.Languages {
					lang.BadgeUrl = BadgeURL(lang.Name)
				}

				data.FetchSpeedStr = FormatTime(data.FetchSpeed)
				data.AnalysisSpeeStr = FormatTime(data.AnalysisSpeed)

				c.JSON(http.StatusOK, TaskInfo{
					Status:       task.Status,
					Done:         task.Status == tasks.STATUS_DONE,
					Error:        false,
					ErrorMessage: "",
					Result:       data,
				})
				return
			case "text/html":
				c.HTML(http.StatusOK, "table.html", data)
				return
			default:
				c.Error(NewTaskStatusError(task, http.StatusBadRequest, "Bad Accept Header"))
			}

		default:
			c.Error(NewTaskStatusError(task, http.StatusNotFound, "Unknown action"))
		}

	}
}
