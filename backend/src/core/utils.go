package core

import (
	"io/fs"
	"os"
	"path"
	"sync"

	q "github.com/golang-collections/collections/queue"
)

func _dirWorker(dir string, fileJobs chan fs.File, wg *sync.WaitGroup, validator *repoValidator) {
	queue := q.New()
	queue.Enqueue(dir)

	for queue.Len() > 0 {
		currentDir := queue.Dequeue().(string)
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return
		}

		for _, entry := range entries {
			entrypath := path.Join(currentDir, entry.Name())
			if entry.IsDir() {
				queue.Enqueue(entrypath)
			} else {
				file, err := os.Open(entrypath)

				if err != nil {
					continue
				}

				if validator.validateFile(entry) {
					wg.Add(1)
					fileJobs <- file
				}

			}
		}
	}

}

func _fileWorker(files chan fs.File, rca *RepositoryAnalyzer, wg *sync.WaitGroup, mu *sync.Mutex) {
	for file := range files {
		rca.addFile(file, mu)

		file.Close()
		wg.Done()
	}
}
