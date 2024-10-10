package api

import (
	"git-analyzer/pkg/tasks"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AnalyzeError struct {
	StatusCode int
	Message    string
}

func (e *AnalyzeError) Error() string {
	return e.Message
}

func NewAnalyzeError(statusCode int, message string) *AnalyzeError {
	return &AnalyzeError{
		StatusCode: statusCode,
		Message:    message,
	}
}

type TaskStatusError struct {
	StatusCode int
	Task       *tasks.RepoTask
	Message    string
}

func (e *TaskStatusError) Error() string {
	return e.Message
}

func NewTaskStatusError(task *tasks.RepoTask, statusCode int, message string) *TaskStatusError {
	return &TaskStatusError{
		StatusCode: statusCode,
		Task:       task,
		Message:    message,
	}
}

func ErrorHandlerForAnalysingRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, err := range c.Errors {
			switch e := err.Err.(type) {
			case *AnalyzeError:
				c.JSON(e.StatusCode, gin.H{
					"id":            "",
					"error":         true,
					"error_message": e.Error(),
					"cache":         false,
					"cache_key":     "",
				})
			case *TaskStatusError:
				if e.Task != nil {
					c.JSON(e.StatusCode, JSONTaskStatus{
						TaskStatus:       e.Task.Status,
						TaskDone:         e.Task.Status == tasks.STATUS_DONE,
						TaskError:        true,
						TaskErrorMessage: e.Error(),
					})
				} else {
					c.JSON(e.StatusCode, JSONTaskStatus{
						TaskStatus:       tasks.STATUS_DONE,
						TaskDone:         true,
						TaskError:        true,
						TaskErrorMessage: e.Error(),
					})
				}
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":         true,
					"error_message": "Something went wrong, please try again later",
				})
			}
		}

	}
}
