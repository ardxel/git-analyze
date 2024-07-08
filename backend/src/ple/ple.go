package ple

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// json format
type language struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

type languageTable struct {
	// extenstion name -> language
	table map[string]string
}

func new() languageTable {
	return languageTable{
		table: make(map[string]string),
	}
}

func (ple *languageTable) set(k, v string) {
	if _, ok := ple.table[k]; !ok {
		ple.table[k] = v
	}
}

func (ple languageTable) Get(ext string) string {
	return ple.table[ext]
}

/*
parse json file with language data
and transform it to simple getter for determining what is the language
used by this extension
for example:
json => { name: "JavaScipt", extensions: ["js", "jsx"] }
parsing json and initializing ple table
lets say we have ".js" extension and we want to know what is the language
ple.Get(".js") -> "JavaScript"
*/
var LT languageTable
var once sync.Once

func _initializeLanguageTable() {
	rootDir, _ := os.Getwd()
	jsonFilePath := rootDir + "/ple.json"
	jsonPLE, err := os.Open(jsonFilePath)
	defer jsonPLE.Close()

	if err != nil {
		fmt.Println(err)
		return
	}

	byteValue, _ := io.ReadAll(jsonPLE)
	pleList := []language{}
	json.Unmarshal(byteValue, &pleList)

	LT = new()

	for _, entity := range pleList {
		for _, ext := range entity.Extensions {
			LT.set(ext, entity.Name)
		}
	}
}

func init() {
	once.Do(_initializeLanguageTable)
}
