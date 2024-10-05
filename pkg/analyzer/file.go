package analyzer

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileInfo struct {
	Name     string
	Files    int32
	Lines    int32
	Blank    int32
	Comments int32
}

func (this *FileInfo) onFile() {
	this.Files++
}

func (this *FileInfo) onLine() {
	this.Lines++
}

func (this *FileInfo) onBlank() {
	this.Blank++
}

func (this *FileInfo) onComment() {
	this.Comments++
}

func Reader(path string) *FileInfo {
	file, _ := os.Open(path)
	base := filepath.Base(file.Name())
	ext := filepath.Ext(base)

	fileInfo := &FileInfo{
		Name:     registry.GetLangByExt(ext),
		Files:    0,
		Lines:    0,
		Blank:    0,
		Comments: 0,
	}

	firstLine := true
	inComm := false

	iterateFileLines(file, func(line string, index int) {
		line = strings.TrimSpace(line)
		fileInfo.onLine()

		if firstLine {
			if strings.HasPrefix(line, "#!") {
				if extByShebang, ok := GetExtByShebang(line); ok {
					fileInfo.Name = registry.GetLangByExt(extByShebang)
				}
			}

			firstLine = false
			fileInfo.onFile()
		}

		if len(line) == 0 {
			fileInfo.onBlank()
		}

		blockComments := registry.GetBlockComments(fileInfo.Name)

		if inComm {
			fileInfo.onComment()

			for _, blockCommPair := range blockComments {
				if strings.HasSuffix(line, blockCommPair[1]) {
					inComm = false
					break
				}
			}
			// skip line comment and block begin comment
			// because we in block comment or exited already
			return
		}

		// match line comment
		for _, lineComm := range registry.GetLineComments(fileInfo.Name) {
			if strings.HasPrefix(line, lineComm) {
				fileInfo.onComment()
			}
		}

		// match begin block comment
		for _, blockCommPair := range blockComments {
			if strings.HasPrefix(line, blockCommPair[0]) {
				inComm = true
				fileInfo.onComment()

				// if block comment started and ended in same line
				if strings.HasSuffix(line[len(blockCommPair[0]):], blockCommPair[1]) {
					inComm = false
				}
			}
		}

	})

	return fileInfo
}

var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func iterateFileLines(file *os.File, onScan func(string, int)) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	defer buf.Reset()

	s := bufio.NewScanner(file)
	s.Buffer(buf.Bytes(), 1024*1024)
	index := 0

	for s.Scan() {
		onScan(s.Text(), index)
		index++
	}
}
