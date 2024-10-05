package analyzer

import (
	"context"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SHEBANG_REGEX = "^#!(.*)"
)

type FileTask struct {
	Path     string          // file path
	Analyzer *RepoAnalyzer   // repository analyzer
	Wg       *sync.WaitGroup // file task can be executed in parallel, so we need to wait for completion
}

func (this *FileTask) Process() {
	defer this.Wg.Done()
	this.Analyzer.AnalyzeFile(this.Path)
}

type Options struct {
	ExcludeFilePatterns []string
	ExcludeDirPatterns  []string
}

var defaultOptions = &Options{
	ExcludeDirPatterns: []string{
		".git",          // Stores version control history
		"node_modules",  // Directory for Node.js dependencies
		".idea",         // Project settings for JetBrains IDEs
		".vscode",       // Project settings for Visual Studio Code
		".venv",         // Virtual environments for Python
		".gradle",       // Gradle build caches and artifacts
		"__pycache__",   // Cached Python bytecode files
		".mypy_cache",   // MyPy type-checker cache for Python
		".pytest_cache", // pytest cache for Python
		".tox",          // Environments for Tox (Python testing tool)
	},
	ExcludeFilePatterns: []string{
		"package-lock.json", // Lockfile for Node.js dependencies
		"yarn.lock",         // Lockfile for Yarn dependencies
		"pipfile.lock",      // Lockfile for Python dependencies
		"Gemfile.lock",      // Lockfile for Ruby dependencies
		"composer.lock",     // Lockfile for PHP dependencies
		"Cargo.lock",        // Lockfile for Rust dependencies
		"*.log",             // Log files
		"*.tmp",             // Temporary files
		"*.swp",             // Swap files created by Vim
		"*.swo",             // Additional swap files created by Vim
		"*.iml",             // Module files for IntelliJ IDEA
		".DS_Store",         // macOS directory metadata
		"thumbs.db",         // Windows thumbnail cache file
		"*.class",           // Compiled Java class files
		"*.pyc",             // Compiled Python bytecode files
		"*.pyo",             // Optimized compiled Python files
		"*.lock",            // General lockfiles from various systems
	},
}

type Result struct {
	TotalFiles    int32       `json:"total_files"`
	TotalLines    int32       `json:"total_lines"`
	TotalBlank    int32       `json:"total_blank"`
	TotalComments int32       `json:"total_comments"`
	Languages     []*Language `json:"languages"`
}

type RepoAnalyzer struct {
	tasks     chan *FileTask
	ctx       context.Context
	cancel    context.CancelFunc
	parallel  bool
	opts      *Options
	languages map[string]*Language
}

func New(opts *Options) *RepoAnalyzer {

	opts.ExcludeDirPatterns = append(opts.ExcludeDirPatterns, defaultOptions.ExcludeDirPatterns...)
	opts.ExcludeFilePatterns = append(opts.ExcludeFilePatterns, defaultOptions.ExcludeFilePatterns...)

	analyzer := &RepoAnalyzer{
		languages: DefinedLanguages(),
		opts:      opts,
	}

	return analyzer
}

func (this *RepoAnalyzer) Result() *Result {
	n := len(this.languages)
	langs := make([]*Language, 0, n)
	total := NewLanguage("TOTAL")

	for _, lang := range this.languages {
		// ignore Total and empty languages
		if lang.Files > 0 && lang.Name != "TOTAL" {
			total.Files += lang.Files
			total.Blank += lang.Blank
			total.Lines += lang.Lines
			total.Comments += lang.Comments
			langs = append(langs, lang)
		}
	}

	result := &Result{
		TotalFiles:    total.Files,
		TotalLines:    total.Lines,
		TotalBlank:    total.Blank,
		TotalComments: total.Comments,
		Languages:     langs,
	}

	// sort by lines,
	// if lines are equal, sort by name,
	sort.SliceStable(result.Languages, func(i, j int) bool {
		lang1, lang2 := result.Languages[i], result.Languages[j]

		if lang1.Lines == lang2.Lines {
			return lang1.Name < lang2.Name
		}

		return lang1.Lines > lang2.Lines
	})

	return result
}

func (this *RepoAnalyzer) AnalyzeFile(path string) {
	result := Reader(path)

	lang := this.languages[result.Name]
	atomic.AddInt32(&lang.Files, result.Files)
	atomic.AddInt32(&lang.Lines, result.Lines)
	atomic.AddInt32(&lang.Blank, result.Blank)
	atomic.AddInt32(&lang.Comments, result.Comments)
}

func (this *RepoAnalyzer) Do(path string, parallelMode bool) (*Result, time.Duration, error) {
	wg := &sync.WaitGroup{}
	if parallelMode {

		this.tasks = make(chan *FileTask, 100)
		this.ctx, this.cancel = context.WithCancel(context.Background())
		defer this.cancel()

		for range runtime.NumCPU() {
			go FileWorker(this.tasks, this.ctx)
		}
	}

	analyzeTimeStart := time.Now()

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()

		if info.IsDir() {
			for _, pattern := range this.opts.ExcludeDirPatterns {
				matched, err := filepath.Match(pattern, name)

				if err != nil || matched {
					return filepath.SkipDir
				}
			}
		}

		for _, pattern := range this.opts.ExcludeFilePatterns {
			matched, _ := filepath.Match(pattern, name)

			if matched || name == pattern {
				return nil
			}
		}

		// if file is bigger than 20KB, use FileTask
		if parallelMode && (info.Size()/1024) > 20 {
			task := &FileTask{
				Path:     path,
				Analyzer: this,
				Wg:       wg,
			}

			select {
			case this.tasks <- task:
				wg.Add(1)
				return nil
			default:
				this.AnalyzeFile(path)
				return nil
			}
		}

		this.AnalyzeFile(path)

		return nil
	})

	wg.Wait()

	analyzeTimeEnd := time.Since(analyzeTimeStart)

	if parallelMode {
		close(this.tasks)
	}

	return this.Result(), analyzeTimeEnd, err
}

func FileWorker(tasks <-chan *FileTask, done context.Context) {
	for {
		select {
		case task := <-tasks:
			if task != nil {
				task.Process()
			}
		case <-done.Done():
			return
		}
	}
}
