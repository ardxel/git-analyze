package core

import (
	"bufio"
	"git-analyzer/src/ple"
	"io/fs"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/go-git/go-git/v5"
)

type RepositoryAnalyzer struct {
	validator   *repoValidator
	langEntries map[string]*LanguageEntry
}

type RepoAnalyzationResult struct {
	TotalFiles  int64                     `json:"total_files"`
	TotalLines  int64                     `json:"total_lines"`
	TotalLangs  int64                     `json:"total_langs"`
	LanguageMap map[string]*LanguageEntry `json:"language_map"`
}

func newRA(validateOptions *RepoValidateOptions) *RepositoryAnalyzer {
	ra := &RepositoryAnalyzer{
		langEntries: make(map[string]*LanguageEntry),
		validator:   createValidator(validateOptions),
	}
	ra.langEntries["TOTAL"] = newLangEntry()

	return ra
}

func (this *RepositoryAnalyzer) Result() *RepoAnalyzationResult {
	result := &RepoAnalyzationResult{}
	result.TotalFiles = this.langEntries["TOTAL"].Files
	result.TotalLines = this.langEntries["TOTAL"].Lines

	delete(this.langEntries, "TOTAL")
	result.TotalLangs = int64(len(this.langEntries))
	result.LanguageMap = this.langEntries

	return result
}

func (this *RepositoryAnalyzer) addFile(file fs.File, mu *sync.Mutex) {
	info, _ := file.Stat()
	name := info.Name()
	ext := path.Ext(name)
	languageName := ple.LT.Get(ext)

	scanner := bufio.NewScanner(file)
	var lines int64
	for scanner.Scan() {
		lines++
	}

	mu.Lock()
	this.langEntries["TOTAL"].Files++
	this.langEntries["TOTAL"].Lines += lines

	entry, ok := this.langEntries[languageName]

	if !ok {
		this.langEntries[languageName] = newLangEntry()
		entry = this.langEntries[languageName]
	}

	entry.Files++
	entry.Lines += lines

	mu.Unlock()
}

func AnalyzeRepository(url string, options *RepoValidateOptions) (*RepoAnalyzationResult, error) {
	dir, _ := os.MkdirTemp("", "git")

	_, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		os.Remove(dir)
		return nil, err
	}

	ra := newRA(options)
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	fileJobs := make(chan fs.File, 10)

	for i := 0; i < runtime.NumCPU(); i++ {
		go _fileWorker(fileJobs, ra, wg, mu)
	}

	_dirWorker(dir, fileJobs, wg, ra.validator)
	wg.Wait()
	close(fileJobs)
	os.Remove(dir)

	return ra.Result(), nil
}
