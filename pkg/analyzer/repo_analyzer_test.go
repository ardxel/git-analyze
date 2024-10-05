package analyzer

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
)

var (
	smallRepoDir string
)

func BenchmarkPrepare(b *testing.B) {
	b.Setenv("USE_FILE_TASKS", "1")

	smallRepo := "https://github.com/ardxel/pet-project-chat"

	dir, err := os.MkdirTemp("", "git")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir) // Ensure the directory is removed after the test

	tm := time.Now()
	_, err = gogit.PlainClone(dir, false, &gogit.CloneOptions{
		Depth: 1,
		URL:   smallRepo,
	})

	smallRepoDir = dir

	println("Time to clone", time.Since(tm).Milliseconds())

	if err != nil {
		b.Fatalf("Failed to clone repository: %v", err)
	}

}

var resultParallel *Result

func BenchmarkAnalyzeRepositoryParallel(b *testing.B) {
	analyzer := New(&Options{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resultParallel, _, _ = analyzer.Do(smallRepoDir, true)
	}
	b.StopTimer()

}

var resultSync *Result

func BenchmarkAnalyzeRepositorySync(b *testing.B) {

	b.ResetTimer()
	analyzer := New(&Options{})
	for i := 0; i < b.N; i++ {
		resultSync, _, _ = analyzer.Do(smallRepoDir, false)
	}
	b.StopTimer()

	fmt.Println("Results are the same:", reflect.DeepEqual(resultParallel, resultSync))
}

var inners = [3][]byte{
	[]byte(`# This is a single-line comment in Python

"""
 This is a block comment in Python
 Also called a multi-line string if not used as a comment
"""

# Function to multiply two numbers
def multiply(a, b):
    return a * b  # Return the product of a and b

# Call the function
result = multiply(5, 3)
print(result)  # Output: 15`),
	[]byte(`// This is a single-line comment in JavaScript

/*
 This is a block comment in JavaScript
 It can span multiple lines
*/

// Function to add two numbers
function add(a, b) {
    return a + b; // Return the sum of a and b
}

// Call the function
let result = add(5, 3);
console.log(result); // Output: 8`),
	[]byte(`// This is a single-line comment in Go

/*
 This is a block comment in Go
 It can span multiple lines
*/

package main

import "fmt"

// Function to subtract two numbers
func subtract(a int, b int) int {
    return a - b // Return the difference between a and b
}

func main() {
    // Call the function
    result := subtract(5, 3)
    fmt.Println(result) // Output: 2
}`),
}

func TestAnalyzeRepositoryParallel(t *testing.T) {
	dir, _ := os.MkdirTemp("", "test")
	defer os.RemoveAll(dir)

	exts := [3]string{".py", ".js", ".go"}

	for i, inner := range inners {
		file, err := os.CreateTemp(dir, "*"+exts[i])
		file.Write(inner)

		if err != nil {
			t.Error(err.Error())
		}
	}

	analyzer := New(&Options{})
	result, _, _ := analyzer.Do(dir, true)

	if result.TotalFiles != 3 {
		t.Errorf("Expected 3 languages, got %d", result.TotalFiles)
	}
	if result.TotalLines != 50 {
		t.Errorf("Expected 50 lines, got %d", result.TotalLines)
	}
	if result.TotalBlank != 11 {
		t.Errorf("Expected 11 blank lines, got %d", result.TotalBlank)
	}
	if result.TotalComments != 21 {
		t.Errorf("Expected 21 comments, got %d", result.TotalComments)
	}
	if len(result.Languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(result.Languages))
	}
}

func TestAnalyzeRepositorySync(t *testing.T) {
	dir, _ := os.MkdirTemp("", "test")
	defer os.RemoveAll(dir)

	exts := [3]string{".py", ".js", ".go"}

	for i, inner := range inners {
		file, err := os.CreateTemp(dir, "*"+exts[i])
		file.Write(inner)

		if err != nil {
			t.Error(err.Error())
		}
	}

	analyzer := New(&Options{})
	result, _, _ := analyzer.Do(dir, false)

	if result.TotalFiles != 3 {
		t.Errorf("Expected 3 languages, got %d", result.TotalFiles)
	}
	if result.TotalLines != 50 {
		t.Errorf("Expected 50 lines, got %d", result.TotalLines)
	}
	if result.TotalBlank != 11 {
		t.Errorf("Expected 11 blank lines, got %d", result.TotalBlank)
	}
	if result.TotalComments != 21 {
		t.Errorf("Expected 21 comments, got %d", result.TotalComments)
	}
	if len(result.Languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(result.Languages))
	}
}
