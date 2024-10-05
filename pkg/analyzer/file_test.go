package analyzer

import (
	"os"
	"testing"
)

func TestAnalyzePythonFile(t *testing.T) {
	inner := []byte(`# This program calculates the sum of two numbers entered by the user

# Step 1: Request the first number from the user
a = int(input("Enter the first number: "))

# Step 2: Request the second number from the user
b = int(input("Enter the second number: "))

"""
Step 3: Add the two numbers
The result is stored in the variable 'sum'
"""
sum = a + b

# Step 4: Display the result
print("The sum of the numbers is:", sum)`)

	file, _ := os.CreateTemp("", "*.py")
	defer os.Remove(file.Name())
	file.Write(inner)
	result := Reader(file.Name())

	if result.Name != "Python" {
		t.Errorf("Expected Python, got %s", result.Name)
	}

	if result.Lines != 16 {
		t.Errorf("PY. Expected 16 lines, got %d", result.Lines)
	}

	if result.Blank != 4 {
		t.Errorf("PY. Expected 4 blank lines, got %d", result.Blank)
	}
	if result.Files != 1 {
		t.Errorf("PY. Expected 1 file, got %d", result.Files)
	}
	if result.Comments != 8 {
		t.Errorf("PY. Expected 8 comments, got %d", result.Comments)
	}
}

func TestAnalyzeJavaScriptFile(t *testing.T) {
	inner := []byte(`// This program checks if a number entered by the user is even or odd

// Step 1: Request a number from the user
let number: number = parseInt(prompt("Enter a number:") || "0");

/*
Step 2: Check if the number is even or odd
- A number is even if it is divisible by 2 with no remainder
- If the remainder is 0, the number is even
- Otherwise, the number is odd
*/
if (number % 2 === 0) {
    console.log("The number is even.");
} else {
    console.log("The number is odd.");
}`)

	file, _ := os.CreateTemp("", "*.js")
	defer os.Remove(file.Name())
	file.Write(inner)
	result := Reader(file.Name())

	if result.Lines != 16 {
		t.Errorf("JS. Expected 16 lines, got %d", result.Lines)
	}
	if result.Blank != 2 {
		t.Errorf("JS. Expected 2 blank lines, got %d", result.Blank)
	}
	if result.Files != 1 {
		t.Errorf("JS. Expected 1 file, got %d", result.Files)
	}
	if result.Comments != 8 {
		t.Errorf("JS. Expected 8 comments, got %d", result.Comments)
	}
}

func TestAnalyzeGoFile(t *testing.T) {
	inner := []byte(`package main

import "fmt"

func main() {
	// Step 1: Declare a variable to store the number
	var number int

	// Step 2: Request a number from the user
	fmt.Print("Enter a number: ")
	fmt.Scan(&number)

	/*
	Step 3: Check if the number is positive, negative, or zero
	- If the number is greater than 0, it's positive
	- If the number is less than 0, it's negative
	- If the number is 0, it's neither positive nor negative
	*/
	if number > 0 {
		fmt.Println("The number is positive.")
	} else if number < 0 {
		fmt.Println("The number is negative.")
	} else {
		fmt.Println("The number is zero.")
	}
}`)

	file, _ := os.CreateTemp("", "*.go")
	defer os.Remove(file.Name())
	file.Write(inner)
	result := Reader(file.Name())

	if result.Lines != 26 {
		t.Errorf("GO. Expected 26 lines, got %d", result.Lines)
	}
	if result.Blank != 4 {
		t.Errorf("GO. Expected 4 blank lines, got %d", result.Blank)
	}
	if result.Files != 1 {
		t.Errorf("GO. Expected 1 file, got %d", result.Files)
	}
	if result.Comments != 8 {
		t.Errorf("GO. Expected 8 comments, got %d", result.Comments)
	}
}

func TestAnalyzeBashFile(t *testing.T) {
	inner := []byte(`#!/bin/bash

# This script checks if a directory exists and creates it if it doesn't

# Step 1: Define the directory name
DIR_NAME="my_directory"

# Step 2: Check if the directory exists
if [ -d "$DIR_NAME" ]; then
    # If the directory exists, print a message
    echo "Directory '$DIR_NAME' already exists."
else
    # If the directory does not exist, create it
    mkdir "$DIR_NAME"
    echo "Directory '$DIR_NAME' has been created."
fi`)

	file, _ := os.CreateTemp("", "bashtest")
	defer os.Remove(file.Name())
	file.Write(inner)
	result := Reader(file.Name())

	if result.Name != "Bash" {
		t.Errorf("Bash. Expected Bash, got %s", result.Name)
	}
	if result.Lines != 16 {
		t.Errorf("Bash. Expected 16 lines, got %d", result.Lines)
	}
	if result.Blank != 3 {
		t.Errorf("Bash. Expected 3 blank lines, got %d", result.Blank)
	}
	if result.Files != 1 {
		t.Errorf("Bash. Expected 1 file, got %d", result.Files)
	}
	if result.Comments != 6 {
		t.Errorf("Bash. Expected 6 comments, got %d", result.Comments)
	}
}
