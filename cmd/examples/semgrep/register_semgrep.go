package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/internal/scanners/semgrep"
)

// SemgrepResult represents a simplified structure of Semgrep results
type SemgrepResult struct {
	Results []struct {
		Path        string `json:"path"`
		Line        int    `json:"start_line"`
		EndLine     int    `json:"end_line"`
		Message     string `json:"message"`
		RuleID      string `json:"rule_id"`
		Severity    string `json:"severity"`
		CodeSnippet string `json:"extra_lines"`
	} `json:"results"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func main() {
	registry := scanners.NewScannerRegistry()

	semgrepScanner := semgrep.NewSemgrepScanner()

	fmt.Println("Setting up Semgrep scanner...")
	if err := semgrepScanner.Setup(); err != nil {
		log.Fatalf("Failed to setup Semgrep scanner: %v", err)
	}

	registry.Register(semgrepScanner)

	if !semgrepScanner.IsInstalled() {
		log.Fatal("Semgrep installation could not be verified")
	}

	fmt.Printf("Scanner: %s\n", semgrepScanner.GetConfig().Name)
	fmt.Printf("Version: %s\n", semgrepScanner.GetInstallationState().Version)
	fmt.Printf("Path: %s\n", semgrepScanner.GetConfig().ExecutablePath)

	// Create a demo file with code issues for demonstration
	tempDir, err := createDemoFiles()
	if err != nil {
		log.Fatalf("Failed to create demo files: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("\nRunning Semgrep scan on demo code at %s...\n", tempDir)

	result, err := semgrepScanner.Scan(tempDir)
	if err != nil {
		fmt.Printf("Scan encountered errors: %v\n", err)
	}

	// Display results
	fmt.Println("\nScan Results:")

	if len(result.Data) > 0 {
		// Try to parse the JSON data
		var semgrepResults SemgrepResult

		// Parse the first line that looks like valid JSON
		for _, line := range result.Data {
			if strings.HasPrefix(line, "{") {
				if err := json.Unmarshal([]byte(line), &semgrepResults); err == nil {
					break
				}
			}
		}

		// Display findings
		if len(semgrepResults.Results) > 0 {
			fmt.Printf("Found %d issues:\n\n", len(semgrepResults.Results))

			for i, finding := range semgrepResults.Results {
				fmt.Printf(" %d. [%s] %s\n", i+1, finding.Severity, finding.Message)
				fmt.Printf("    File: %s (lines %d-%d)\n", finding.Path, finding.Line, finding.EndLine)
				fmt.Printf("    Rule: %s\n\n", finding.RuleID)
			}
		} else {
			fmt.Println("No issues found")
		}

		// Display any errors from Semgrep
		if len(semgrepResults.Errors) > 0 {
			fmt.Println("\nSemgrep Errors:")
			for _, err := range semgrepResults.Errors {
				fmt.Printf(" - %s\n", err.Message)
			}
		}
	} else {
		fmt.Println("No output from scan")
	}

	fmt.Println("\nScan completed successfully")
}

// Create temporary files with code issues for demonstration
func createDemoFiles() (string, error) {
	// Use os.MkdirTemp instead of ioutil.TempDir
	tempDir, err := os.MkdirTemp("", "semgrep-demo-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create a Python file with security issues
	pythonFile := filepath.Join(tempDir, "demo.py")
	pythonContent := `
import os
import pickle
import subprocess

def insecure_function(user_input):
    # Command injection vulnerability
    os.system("ls " + user_input)

    # Unsafe deserialization
    data = pickle.loads(user_input)

    # Hardcoded credentials
    password = "super_secret_password123"
    api_key = "1234567890abcdef1234567890abcdef"

    return data

# SQL Injection vulnerability
def query_db(user_input):
    query = "SELECT * FROM users WHERE username = '" + user_input + "'"
    return execute_query(query)

def execute_query(q):
    pass
`

	// Use os.WriteFile instead of ioutil.WriteFile
	if err := os.WriteFile(pythonFile, []byte(pythonContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create Python test file: %w", err)
	}

	// Create a JavaScript file with security issues
	jsFile := filepath.Join(tempDir, "demo.js")
	jsContent := `
const express = require('express');
const app = express();

app.get('/search', (req, res) => {
  // XSS vulnerability
  res.send('<div>' + req.query.term + '</div>');

  // Prototype pollution
  const data = JSON.parse(req.query.data);
  Object.assign({}, data);

  // Insecure eval
  eval(req.query.code);
});

// Hardcoded JWT secret
const jwt_secret = "very_secret_key_do_not_share";

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
`

	// Use os.WriteFile instead of ioutil.WriteFile
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create JavaScript test file: %w", err)
	}

	return tempDir, nil
}
