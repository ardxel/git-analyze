package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DiskSize       int64
	MaxRepoSize    int64
	SyncEvery      int32
	UseFileWorkers bool
	Debug          bool
	GoEnv          string
	MainPort       string
	RedisHost      string
	RedisPort      string
	GithubApiPat   string
}

var Vars *Config

type Task interface {
	Process()
	Done() any
}

func isTest() bool {
	envTest, ok := os.LookupEnv("GO_ENV")

	if !ok {
		panic("Missing GO_ENV")
	}

	return envTest == "test"
}

func getEnv(key string) string {
	env, ok := os.LookupEnv(key)

	// if running in test env, allow to override
	if !ok && isTest() {
		return ""
	}

	if !ok {
		panic(fmt.Errorf("Missing env var %s", key))
	}

	return env
}

func getEnvInt(key string) int {
	env := getEnv(key)
	val, err := strconv.Atoi(env)

	if err != nil {
		if isTest() {
			return 0
		}
		panic(fmt.Errorf("Invalid env var %s", key))
	}

	return val
}

func getEnvBool(key string) bool {
	env := getEnv(key)

	if env == "1" || env == "true" {
		return true
	}

	return false
}

var onceInitConfig = &sync.Once{}

func InitConfig() {
	Vars = &Config{
		DiskSize:       int64(getEnvInt("MAX_DISK_SIZE")),
		MaxRepoSize:    int64(getEnvInt("MAX_REPO_SIZE")),
		SyncEvery:      int32(getEnvInt("SYNC_EVERY")),
		UseFileWorkers: getEnvBool("USE_FILE_WORKERS"),
		Debug:          getEnvBool("DEBUG"),
		MainPort:       getEnv("MAIN_PORT"),
		RedisPort:      getEnv("REDIS_PORT"),
		RedisHost:      getEnv("REDIS_HOST"),
		GoEnv:          getEnv("GO_ENV"),
		GithubApiPat:   getEnv("API_PAT"),
	}
}

func init() {
	onceInitConfig.Do(InitConfig)
}
