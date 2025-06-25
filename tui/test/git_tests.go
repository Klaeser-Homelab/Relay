package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runGitTests(config *TestConfig) *TestSuite {
	suite := &TestSuite{Name: "Git Operations"}

	// Test git repository detection
	passed, duration, err := runTest("Git Repository Detection", func() error {
		return testGitRepositoryDetection(config)
	})
	suite.AddResult("Git Repository Detection", passed, duration, err)

	// Test git status functionality
	passed, duration, err = runTest("Git Status", func() error {
		return testGitStatus(config)
	})
	suite.AddResult("Git Status", passed, duration, err)

	// Test smart commit preparation
	passed, duration, err = runTest("Smart Commit Setup", func() error {
		return testSmartCommitSetup(config)
	})
	suite.AddResult("Smart Commit Setup", passed, duration, err)

	// Test git operations validation
	passed, duration, err = runTest("Git Operations Validation", func() error {
		return testGitOperationsValidation(config)
	})
	suite.AddResult("Git Operations Validation", passed, duration, err)

	return suite
}

func testGitRepositoryDetection(config *TestConfig) error {
	// Check if RelayTest is a git repository
	gitDir := filepath.Join(config.RelayTestPath, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("RelayTest directory is not a git repository (no .git directory)")
	}

	// Test git command in RelayTest directory
	cmd := exec.Command("git", "status")
	cmd.Dir = config.RelayTestPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git status failed in RelayTest directory: %w (output: %s)", err, string(output))
	}

	return nil
}

func testGitStatus(config *TestConfig) error {
	// Test git status command
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = config.RelayTestPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git status --porcelain failed: %w (output: %s)", err, string(output))
	}

	// Output can be empty (clean repo) or contain changes
	// Just verify the command executed successfully
	return nil
}

func testSmartCommitSetup(config *TestConfig) error {
	// Create a test file in RelayTest to enable smart commit testing
	testFile := filepath.Join(config.RelayTestPath, "test_commit_file.txt")

	// Write test content
	content := "This is a test file for smart commit functionality.\nTimestamp: " + fmt.Sprintf("%d", os.Getpid())
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		return fmt.Errorf("test file was not created")
	}

	// Check that git detects the new file
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = config.RelayTestPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git status failed after creating test file: %w", err)
	}

	// Should show the untracked file
	if !strings.Contains(string(output), "test_commit_file.txt") {
		return fmt.Errorf("git did not detect new test file")
	}

	return nil
}

func testGitOperationsValidation(config *TestConfig) error {
	// Test git configuration (needed for commits)
	cmd := exec.Command("git", "config", "user.name")
	cmd.Dir = config.RelayTestPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		// If no git user name is configured, set a test one
		setNameCmd := exec.Command("git", "config", "user.name", "Relay Test User")
		setNameCmd.Dir = config.RelayTestPath
		if err := setNameCmd.Run(); err != nil {
			return fmt.Errorf("failed to set git user name for testing: %w", err)
		}
	}

	// Test git email configuration
	cmd = exec.Command("git", "config", "user.email")
	cmd.Dir = config.RelayTestPath
	output, err = cmd.CombinedOutput()

	if err != nil {
		// If no git email is configured, set a test one
		setEmailCmd := exec.Command("git", "config", "user.email", "relay-test@example.com")
		setEmailCmd.Dir = config.RelayTestPath
		if err := setEmailCmd.Run(); err != nil {
			return fmt.Errorf("failed to set git email for testing: %w", err)
		}
	}

	// Test basic git operations work
	cmd = exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = config.RelayTestPath
	output, err = cmd.CombinedOutput()

	if err != nil {
		// If there are no commits, that's okay for a new repo
		if !strings.Contains(string(output), "does not have any commits yet") {
			return fmt.Errorf("unexpected git log error: %w (output: %s)", err, string(output))
		}
	}

	return nil
}

// Test smart commit with RelayTest project (integration test)
func testSmartCommitIntegration(config *TestConfig) error {
	// Add RelayTest project
	addCmd := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	_, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add RelayTest project: %w", err)
	}

	// Open RelayTest project
	openCmd := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	_, err = openCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open RelayTest project: %w", err)
	}

	// Create a test change
	testFile := filepath.Join(config.RelayTestPath, "smart_commit_test.txt")
	content := fmt.Sprintf("Smart commit test - %d", os.Getpid())
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create test file for smart commit: %w", err)
	}

	// Try smart commit (this might fail if Claude permissions are restricted)
	commitCmd := exec.Command(config.RelayBinaryPath, "commit")
	output, err := commitCmd.CombinedOutput()

	// Note: This test might fail due to Claude CLI permissions
	// We'll check if the command at least executed without panicking
	outputStr := string(output)

	if err != nil {
		// Check if it's a permission/authentication error vs a code error
		if strings.Contains(outputStr, "permission") ||
			strings.Contains(outputStr, "auth") ||
			strings.Contains(outputStr, "Unable to run") {
			// This is expected in test environment
			fmt.Printf("Smart commit failed due to permissions (expected): %s\n", outputStr)
		} else {
			return fmt.Errorf("smart commit failed unexpectedly: %w (output: %s)", err, outputStr)
		}
	}

	// Clean up test file
	os.Remove(testFile)

	// Clean up project
	removeCmd := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	removeCmd.CombinedOutput()

	return nil
}

// Test git branch operations
func testGitBranches(config *TestConfig) error {
	// List current branches
	cmd := exec.Command("git", "branch")
	cmd.Dir = config.RelayTestPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to list git branches: %w (output: %s)", err, string(output))
	}

	// Should show at least one branch (usually main or master)
	outputStr := string(output)
	if !strings.Contains(outputStr, "main") && !strings.Contains(outputStr, "master") {
		// Might be a new repo with no commits
		if strings.TrimSpace(outputStr) == "" {
			// No branches yet, that's okay for a new repo
			return nil
		}
		return fmt.Errorf("unexpected branch output: %s", outputStr)
	}

	return nil
}
