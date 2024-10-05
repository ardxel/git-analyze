package analyzer

import (
	"encoding/json"
	"git-analyzer/pkg/config"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type LanguageData struct {
	Name          string     `json:"name"`
	Extensions    []string   `json:"extensions"`
	LineComments  []string   `json:"lineComment"`
	BlockComments [][]string `json:"blockComment"`
}

var registry *LanguageRegistry

type LanguageRegistry struct {
	langsByName   map[string]*LanguageData
	langnameByExt map[string]string
}

func (r *LanguageRegistry) GetLangs() []string {
	langs := make([]string, 0, len(r.langsByName))

	for langname := range r.langsByName {
		langs = append(langs, langname)
	}

	return langs
}

func (r *LanguageRegistry) GetLangByExt(ext string) string {
	if len(ext) == 0 {
		return "Other"
	}

	if ext[0] == '.' {
		ext = ext[1:]
	}

	langName, ok := r.langnameByExt[ext]

	if !ok {
		return "Other"
	}

	return langName
}

func (r *LanguageRegistry) GetLineComments(langName string) []string {
	data, ok := r.langsByName[langName]

	if !ok {
		return []string{}
	}

	return data.LineComments
}

// Получение маркеров блочных комментариев
func (r *LanguageRegistry) GetBlockComments(langName string) [][]string {
	data, ok := r.langsByName[langName]

	if !ok {
		return [][]string{}
	}

	return data.BlockComments
}

func initLanguageRegistry() {
	rootDir, _ := os.Getwd()

	if config.Vars.GoEnv == "test" {
		rootDir = filepath.Join(rootDir, "../../")
	}

	jsonFilePath := filepath.Join(rootDir, "ple.json")
	jsonPLE, err := os.Open(jsonFilePath)
	defer jsonPLE.Close()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
		return
	}

	byteValue, _ := io.ReadAll(jsonPLE)
	langList := []LanguageData{}
	json.Unmarshal(byteValue, &langList)
	registry = &LanguageRegistry{
		langsByName:   make(map[string]*LanguageData),
		langnameByExt: make(map[string]string),
	}

	for _, entity := range langList {
		registry.langsByName[entity.Name] = &entity

		for _, ext := range entity.Extensions {
			registry.langnameByExt[ext] = entity.Name
		}
	}
}

var onceInitRegistry sync.Once

func init() {
	onceInitRegistry.Do(initLanguageRegistry)
}
