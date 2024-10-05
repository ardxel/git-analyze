package analyzer

import (
	"path/filepath"
	"strings"
)

type Language struct {
	Name     string `json:"name" redis:"name"`
	Blank    int32  `json:"blank" redis:"blank"`
	Comments int32  `json:"comments" redis:"comments"`
	Lines    int32  `json:"lines" redis:"lines"`
	Files    int32  `json:"files" redis:"files"`
}

func NewLanguage(name string) *Language {
	return &Language{
		Name:     name,
		Blank:    0,
		Comments: 0,
		Lines:    0,
		Files:    0,
	}
}

func DefinedLanguages() map[string]*Language {
	m := make(map[string]*Language)

	for _, lang := range registry.GetLangs() {
		m[lang] = NewLanguage(lang)
	}

	m["Other"] = NewLanguage("Other")

	return m
}

var shebangExts = map[string]string{
	"node":    "js",
	"python":  "py",
	"python3": "py",
	"perl":    "pl",
	"ruby":    "rb",
	"make":    "make",
	"rc":      "plan9sh",
	"gosh":    "scm",
	"escript": "erl",
}

func GetExtByShebang(line string) (lang string, ok bool) {
	if !strings.HasPrefix(line, "#!") {
		return "", false
	}

	parts := strings.Fields(line[2:])

	if len(parts) == 0 {
		return "", false
	}

	if len(parts) == 2 {
		interpreter := parts[1]

		if ext, ok := shebangExts[parts[1]]; ok {
			interpreter = ext
		}

		return interpreter, true
	}

	if ext, ok := shebangExts[filepath.Base(parts[0])]; ok {
		return ext, ok
	}

	return filepath.Base(parts[0]), true
}
