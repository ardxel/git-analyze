package core

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

type RepoValidateOptions struct {
	ExcludeFilePatterns []string
	ExcludeDirPatterns  []string
}

type repoValidator struct {
	excludeFilePatterns []string
	excludeDirPatterns  []string
}

func createValidator(options *RepoValidateOptions) *repoValidator {
	validator := &repoValidator{}

	if len(options.ExcludeFilePatterns) > 0 {
		validator.excludeFilePatterns = options.ExcludeFilePatterns
	} else {
		validator.excludeFilePatterns = []string{}
	}

	if len(options.ExcludeDirPatterns) > 0 {
		validator.excludeDirPatterns = options.ExcludeDirPatterns
	} else {
		validator.excludeDirPatterns = []string{}
	}

	return validator
}

func (this repoValidator) validateFile(file fs.DirEntry) bool {
	if len(this.excludeFilePatterns) > 0 {
		for _, pattern := range this.excludeFilePatterns {
			matched, err := filepath.Match(pattern, file.Name())

			if err != nil {
				fmt.Println(err)
				return false
			}

			if matched {
				return false
			}
		}
	}

	return true
}
