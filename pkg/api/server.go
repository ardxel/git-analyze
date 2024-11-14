package api

import (
	"context"
	"fmt"
	"git-analyzer/pkg/config"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Redis *RedisDB
}

func New() *Server {
	return &Server{
		Redis: CreateRedisDB(),
	}
}

func (s *Server) CheckCredentials() error {
	ctx := context.Background()
	_, resp, err := githubClient.Users.Get(ctx, "")
	if err != nil && resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Invalid token!")
		// panic("Invalid token!")
	}
	return nil
}

func (s *Server) IsProduction() bool {
	return config.Vars.GoEnv == "production"
}

func (s *Server) ConfigureMiddleware(r *gin.Engine) {
	isProd := s.IsProduction()
	if isProd {
		r.Use(gin.Recovery())
		r.Use(CSP())
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"POST", "GET"},
			AllowCredentials: true,
		}))
	}
	r.Use(gin.Logger())
}

func (s *Server) ConfigureHandlers(r *gin.Engine) {
	r.GET("/", HandleGetForm(s))
	apiGroup := r.Group("/api")

	{
		apiGroup.Use(ErrorHandlerForAnalysingRoutes())

		createTaskHandlers := []gin.HandlerFunc{
			// RedisRateLimitMV(this),
			ValidateFormMV(s),
			// RedisRepoTaskCacheMV(this),
			HandleCreateTask(s),
		}

		apiGroup.POST("/task", createTaskHandlers...)
		apiGroup.GET("/task/:id/:action", HandleTask(s))
	}
}

func (s *Server) ConfigureStatic(r *gin.Engine) {
	isProd := s.IsProduction()
	statics := [4]string{"assets/js", "assets/css", "assets/img", "assets/templates"}

	if isProd {
		for i, path := range statics {
			statics[i] = filepath.Join("dist", filepath.Base(path))
		}
	}

	r.LoadHTMLGlob(fmt.Sprintf("%s/*.html", statics[3]))
	for _, static := range statics[:3] {
		r.Static("/"+filepath.Base(static), static)
	}

}

func (s *Server) Start() {
	if err := s.CheckCredentials(); err != nil {
		panic(err.Error())
	}

	isProd := s.IsProduction()

	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"FormatTime": FormatTime,
		"BadgeURL":   BadgeURL,
	})

	s.ConfigureMiddleware(r)
	s.ConfigureHandlers(r)
	s.ConfigureStatic(r)

	srvr := &http.Server{
		Addr:    ":" + config.Vars.MainPort,
		Handler: r,
	}

	if err := srvr.ListenAndServe(); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
