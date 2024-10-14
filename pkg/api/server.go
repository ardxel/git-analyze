package api

import (
	"fmt"
	"git-analyzer/pkg/config"
	"log"
	"net/http"
	"text/template"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Redis *RedisDB
}

func New() *Server {
	s := &Server{
		Redis: CreateRedisDB(),
	}

	return s
}

func (s *Server) ConfigureMiddleware(r *gin.Engine) {
	if config.Vars.GoEnv == "development" {
		r.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		})
	}

	if config.Vars.GoEnv == "production" {
		r.Use(gin.Recovery())

		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"POST", "GET"},
			AllowCredentials: true,
		}))

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

		r.Use(func(c *gin.Context) {
			c.Header("Content-Security-Policy", cspValue)
		})
	}

	r.Use(gin.Logger())
}

func (this *Server) ConfigureHandlers(r *gin.Engine) {
	r.GET("/", HandleGetForm(this))

	apiGroup := r.Group("/api")

	{
		apiGroup.Use(ErrorHandlerForAnalysingRoutes())

		createTaskHandlers := []gin.HandlerFunc{
			RedisRateLimitMV(this),
			ValidateFormMV(this),
			RedisRepoTaskCacheMV(this),
			HandleCreateTask(this),
		}

		apiGroup.POST("/task", createTaskHandlers...)
		apiGroup.GET("/task/:id/:action", HandleTask(this))
	}

}

func (this *Server) Start() {
	if config.Vars.GoEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"FormatTime": formatTime,
		"BadgeURL":   badgeURL,
	})

	r.LoadHTMLGlob("templates/*.html")

	this.ConfigureMiddleware(r)
	this.ConfigureHandlers(r)

	jsDir := "./static/js"
	cssDir := "./static/css"
	imgDir := "./static/img"

	if config.Vars.GoEnv == "production" {
		jsDir = "./dist/js"
		cssDir = "./dist/css"
		imgDir = "./dist/img"
	}

	r.Static("/js", jsDir)
	r.Static("/css", cssDir)
	r.Static("/img", imgDir)

	srv := &http.Server{
		Addr:    ":" + config.Vars.MainPort,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
