package tasks

import (
	"errors"
	"fmt"
	"git-analyzer/pkg/analyzer"
	"git-analyzer/pkg/config"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

const (
	REQUIRED_LIMIT    int64  = 104857600 // Minimum required free space: 100 MB in bytes
	TEMP_FILE_PATTERN string = "git"     // repository dir name pattern for better repository identification

	// Repository task statuses
	STATUS_INIT    uint8 = 1
	STATUS_FETCH   uint8 = 2
	STATUS_ANALYZE uint8 = 3
	STATUS_DONE    uint8 = 4
)

var RepoTaskQueue *TaskQueue

// TaskQueue is needed to simultaneously manage disk space,
// rite and analyze repositories using a repository analyzer,
// and also have some kind of queue of client requests.
//
// The server can receive many requests in asynchronous mode,
// but the server will not always be able to analyze the repository due to lack of disk space.
// Each async request will be queued and processed by a single goroutine
type TaskQueue struct {
	maxDiskSize    int64          // max available space on disk in bytes
	MaxRepoSize    int64          // max repository size in bytes
	freeMemory     int64          // free memory in bytes
	TaskChan       chan *RepoTask // channel of repository tasks
	Cache          *ttlcache.Cache[string, *RepoTask]
	useFileWorkers bool

	// all calculations with free memory are not carried out directly with the disk,
	// but only superficially, so it is important to at least sometimes
	// synchronize real data on the filled memory.
	// synchronize every N-th usage
	syncEvery    int32
	writingCount int32  // how many write operations are performed
	rootDir      string // root directory of repositories
}

func (this *TaskQueue) GetTask(id string) (*RepoTask, bool) {
	task := this.Cache.Get(id)

	if task == nil {
		return nil, false
	}

	val := task.Value()

	return val, true
}

func (this *TaskQueue) DeleteTask(id string) {
	this.Cache.Delete(id)
}

func (this *TaskQueue) Add(task *RepoTask) string {
	taskID := uuid.New().String()
	this.Cache.Set(taskID, task, ttlcache.DefaultTTL)
	this.TaskChan <- task
	return taskID
}

// check if there is enough free space to write a repository of the given size in bytes on the disk.
func (this *TaskQueue) canWrite(size int64) bool {

	if config.Vars.Debug {
		log.Printf("total: %d MB, free: %d MB, repo size: %d MB", this.maxDiskSize/1048576, this.freeMemory/1048576, size/1048576)
	}

	// Check if repository size exceeds the maximum allowable size
	if size > this.MaxRepoSize {
		return false
	}

	// Check if free memory is less than the required limit
	if this.freeMemory < REQUIRED_LIMIT {
		return false
	}

	// Check if the new directory size exceeds the maximum allowable size minus the required limit
	if size > this.maxDiskSize-REQUIRED_LIMIT {
		return false
	}

	return true
}

// try to clone a repository in the root folder and return the path of the repository
func (this *TaskQueue) writeRepo(task *RepoTask) (path string, fetchSpeed time.Duration, err error) {
	if !this.canWrite(task.Size) {
		return "", 0, errors.New("Memory limit exceeded")
	}

	dir, _ := os.MkdirTemp("", TEMP_FILE_PATTERN)
	fetchRepoStart := time.Now()

	// clone the repository
	// works only for public repositories
	_, err = gogit.PlainClone(dir, false, &gogit.CloneOptions{
		Depth: 1,
		URL:   task.GetURL(),
	})

	fetchRepoEnd := time.Since(fetchRepoStart)

	if err != nil {
		return "", 0, err
	}

	if config.Vars.Debug {
		log.Printf("Repo cloned in %d ms\n", fetchRepoEnd.Milliseconds())
	}

	this.freeMemory -= task.Size
	this.writingCount++
	this.syncMemory()

	return dir, fetchRepoEnd, nil
}

// sync real memory usage of the disk
func (this *TaskQueue) syncMemory() {
	if this.writingCount < this.syncEvery {
		return
	}

	var usedSpace int64

	// walk through the root dir and count the size of all patterned files
	filepath.WalkDir(this.rootDir, func(path string, e os.DirEntry, err error) error {

		if e.IsDir() {
			// if the directory is located at the 1st level of nesting,
			// directly in the root and is not patterned, skip
			if filepath.Dir(path) == this.rootDir && !strings.Contains(e.Name(), TEMP_FILE_PATTERN) {
				return filepath.SkipDir
			}
			return nil
		}

		fileInfo, _ := e.Info()
		usedSpace += fileInfo.Size()

		return nil
	})

	this.writingCount = 0 // reset usage count
	this.freeMemory = this.maxDiskSize - usedSpace
}

type RepoTask struct {
	Status uint8 // Task status
	// ID            string            // Task ID
	Size          int64             // size of repository in bytes
	Owner         string            // repository owner
	Name          string            // repository name
	Opts          *analyzer.Options // validation options
	Result        *analyzer.Result
	FetchSpeed    time.Duration
	AnalysisSpeed time.Duration
	Err           error
}

func (this *RepoTask) GetURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", this.Owner, this.Name)
}

func (this *RepoTask) UpdateStatus(status uint8) {
	this.Status = status
}

func (this *RepoTask) Process() {
	defer this.UpdateStatus(STATUS_DONE)

	this.UpdateStatus(STATUS_FETCH)

	if !RepoTaskQueue.canWrite(this.Size) {
		this.Result = nil
		this.Err = fmt.Errorf("Memory limit exceeded")
		this.UpdateStatus(STATUS_DONE)
		return
	}

	dir, fetchSpeed, err := RepoTaskQueue.writeRepo(this)

	defer os.RemoveAll(dir)

	if err != nil {
		this.Err = err
		this.Result = nil
		this.UpdateStatus(STATUS_DONE)
		return
	}

	this.UpdateStatus(STATUS_ANALYZE)

	repoAnalyzer := analyzer.New(this.Opts)
	result, analysisSpeed, err := repoAnalyzer.Do(dir, config.Vars.UseFileWorkers)

	this.FetchSpeed = fetchSpeed
	this.AnalysisSpeed = analysisSpeed
	this.Result = result
	this.Err = err
}

func InitMe() {
	cache := ttlcache.New(
		ttlcache.WithTTL[string, *RepoTask](3 * time.Minute),
	)

	go cache.Start()

	var (
		syncEvery          = config.Vars.SyncEvery
		maxDiskSizeInBytes = config.Vars.DiskSize * 1024 * 1024
		maxRepoSize        = config.Vars.MaxRepoSize * 1024 * 1024
	)

	q := &TaskQueue{
		TaskChan:       make(chan *RepoTask, 20),
		useFileWorkers: config.Vars.UseFileWorkers,
		maxDiskSize:    maxDiskSizeInBytes,
		freeMemory:     maxDiskSizeInBytes,
		MaxRepoSize:    maxRepoSize,
		Cache:          cache,
		rootDir:        os.TempDir(),
		syncEvery:      syncEvery,
		writingCount:   syncEvery + 1,
	}

	q.syncMemory() // sync memory for the first time

	// go single goroutine for managing tasks
	go func() {
		for {
			select {
			case task := <-q.TaskChan:
				task.Process()
			}
		}
	}()

	RepoTaskQueue = q
}

var onceInitTaskQueue sync.Once

func init() {
	onceInitTaskQueue.Do(InitMe)
}
