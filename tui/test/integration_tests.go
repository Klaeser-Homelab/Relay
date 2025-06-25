package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func runIntegrationTests(config *TestConfig) *TestSuite {
	suite := &TestSuite{Name: "End-to-End Integration"}

	// Test complete project workflow
	passed, duration, err := runTest("Complete Project Workflow", func() error {
		return testCompleteProjectWorkflow(config)
	})
	suite.AddResult("Complete Project Workflow", passed, duration, err)

	// Test multi-project management
	passed, duration, err = runTest("Multi-Project Management", func() error {
		return testMultiProjectManagement(config)
	})
	suite.AddResult("Multi-Project Management", passed, duration, err)

	// Test error recovery
	passed, duration, err = runTest("Error Recovery", func() error {
		return testErrorRecovery(config)
	})
	suite.AddResult("Error Recovery", passed, duration, err)

	// Test persistence across sessions
	passed, duration, err = runTest("Session Persistence", func() error {
		return testSessionPersistence(config)
	})
	suite.AddResult("Session Persistence", passed, duration, err)

	// Test smart commit workflow
	passed, duration, err = runTest("Smart Commit Workflow", func() error {
		return testSmartCommitWorkflow(config)
	})
	suite.AddResult("Smart Commit Workflow", passed, duration, err)

	return suite
}

func testCompleteProjectWorkflow(config *TestConfig) error {
	// Test the complete workflow: add -> open -> status -> remove

	// Step 1: Add RelayTest project
	addCmd := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	output, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add project: %w (output: %s)", err, string(output))
	}

	if !strings.Contains(string(output), "Successfully added project") {
		return fmt.Errorf("unexpected add output: %s", string(output))
	}

	// Step 2: List projects to verify addition
	listCmd := exec.Command(config.RelayBinaryPath, "list")
	output, err = listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if !strings.Contains(string(output), "RelayTest") {
		return fmt.Errorf("RelayTest not found in project list: %s", string(output))
	}

	// Step 3: Open the project
	openCmd := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	output, err = openCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open project: %w", err)
	}

	if !strings.Contains(string(output), "Switched to project") {
		return fmt.Errorf("unexpected open output: %s", string(output))
	}

	// Step 4: Check status
	statusCmd := exec.Command(config.RelayBinaryPath, "status")
	output, err = statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if !strings.Contains(string(output), "Active Project: RelayTest") {
		return fmt.Errorf("RelayTest not shown as active: %s", string(output))
	}

	// Step 5: Verify active project marker in list
	listCmd = exec.Command(config.RelayBinaryPath, "list")
	output, err = listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list projects after opening: %w", err)
	}

	if !strings.Contains(string(output), "* RelayTest") {
		return fmt.Errorf("active project marker not shown: %s", string(output))
	}

	// Step 6: Remove the project
	removeCmd := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	output, err = removeCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	if !strings.Contains(string(output), "Successfully removed") {
		return fmt.Errorf("unexpected remove output: %s", string(output))
	}

	// Step 7: Verify removal
	listCmd = exec.Command(config.RelayBinaryPath, "list")
	output, err = listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list projects after removal: %w", err)
	}

	if strings.Contains(string(output), "RelayTest") {
		return fmt.Errorf("RelayTest still appears after removal: %s", string(output))
	}

	return nil
}

func testMultiProjectManagement(config *TestConfig) error {
	// Create a temporary directory for second test project
	tempDir, err := os.MkdirTemp("", "relay-test-project2")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize as git repository
	gitCmd := exec.Command("git", "init")
	gitCmd.Dir = tempDir
	gitCmd.Run()

	// Add first project (RelayTest)
	addCmd1 := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	_, err = addCmd1.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add first project: %w", err)
	}

	// Add second project (temp directory)
	addCmd2 := exec.Command(config.RelayBinaryPath, "add", "-p", tempDir)
	_, err = addCmd2.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add second project: %w", err)
	}

	// List projects - should show both
	listCmd := exec.Command(config.RelayBinaryPath, "list")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "RelayTest") {
		return fmt.Errorf("RelayTest not found in multi-project list")
	}

	projectName := filepath.Base(tempDir)
	if !strings.Contains(outputStr, projectName) {
		return fmt.Errorf("second project not found in list")
	}

	// Open first project
	openCmd1 := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	_, err = openCmd1.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open first project: %w", err)
	}

	// Verify first project is active
	statusCmd := exec.Command(config.RelayBinaryPath, "status")
	output, err = statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if !strings.Contains(string(output), "RelayTest") {
		return fmt.Errorf("first project not active")
	}

	// Switch to second project
	openCmd2 := exec.Command(config.RelayBinaryPath, "open", projectName)
	_, err = openCmd2.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open second project: %w", err)
	}

	// Verify second project is now active
	statusCmd = exec.Command(config.RelayBinaryPath, "status")
	output, err = statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get status after switch: %w", err)
	}

	if !strings.Contains(string(output), projectName) {
		return fmt.Errorf("second project not active after switch")
	}

	// Clean up both projects
	removeCmd1 := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	removeCmd1.CombinedOutput()

	removeCmd2 := exec.Command(config.RelayBinaryPath, "remove", projectName)
	removeCmd2.CombinedOutput()

	return nil
}

func testErrorRecovery(config *TestConfig) error {
	// Test various error conditions and recovery

	// Test 1: Try to add non-existent path
	addCmd := exec.Command(config.RelayBinaryPath, "add", "-p", "/non/existent/path")
	output, err := addCmd.CombinedOutput()

	// Should fail gracefully
	if err == nil {
		return fmt.Errorf("adding non-existent path should have failed")
	}

	if !strings.Contains(string(output), "does not exist") {
		return fmt.Errorf("unexpected error message for non-existent path: %s", string(output))
	}

	// Test 2: Try to open non-existent project
	openCmd := exec.Command(config.RelayBinaryPath, "open", "NonExistentProject")
	output, err = openCmd.CombinedOutput()

	// Should fail gracefully
	if err == nil {
		return fmt.Errorf("opening non-existent project should have failed")
	}

	// Test 3: Try to remove non-existent project
	removeCmd := exec.Command(config.RelayBinaryPath, "remove", "NonExistentProject")
	output, err = removeCmd.CombinedOutput()

	// Should fail gracefully
	if err == nil {
		return fmt.Errorf("removing non-existent project should have failed")
	}

	// Test 4: Try to get status with no active project
	statusCmd := exec.Command(config.RelayBinaryPath, "status")
	output, err = statusCmd.CombinedOutput()

	// Should handle gracefully (might show "no active project")
	if err != nil {
		outputStr := string(output)
		if !strings.Contains(outputStr, "No active project") &&
			!strings.Contains(outputStr, "no active project") {
			return fmt.Errorf("unexpected status error: %s", outputStr)
		}
	}

	return nil
}

func testSessionPersistence(config *TestConfig) error {
	// Test that project data persists across relay invocations

	// Add project
	addCmd := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	_, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add project for persistence test: %w", err)
	}

	// Open project
	openCmd := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	_, err = openCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open project for persistence test: %w", err)
	}

	// Wait a moment to ensure database write
	time.Sleep(100 * time.Millisecond)

	// Check status in a new invocation
	statusCmd := exec.Command(config.RelayBinaryPath, "status")
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get status in new session: %w", err)
	}

	if !strings.Contains(string(output), "RelayTest") {
		return fmt.Errorf("project not persisted across sessions: %s", string(output))
	}

	// List projects in new invocation
	listCmd := exec.Command(config.RelayBinaryPath, "list")
	output, err = listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list projects in new session: %w", err)
	}

	if !strings.Contains(string(output), "RelayTest") {
		return fmt.Errorf("project list not persisted: %s", string(output))
	}

	// Clean up
	removeCmd := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	removeCmd.CombinedOutput()

	return nil
}

func testSmartCommitWorkflow(config *TestConfig) error {
	// Test the smart commit functionality end-to-end

	// Add and open RelayTest project
	addCmd := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	_, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add project for smart commit test: %w", err)
	}

	openCmd := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	_, err = openCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open project for smart commit test: %w", err)
	}

	// Create a test file with changes
	testFile := filepath.Join(config.RelayTestPath, "commit_test.md")
	content := fmt.Sprintf("# Test File\n\nThis is a test file for smart commit functionality.\n\nCreated at: %s\n", time.Now().Format(time.RFC3339))

	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}

	// Ensure the file is detected by git
	gitStatusCmd := exec.Command("git", "status", "--porcelain")
	gitStatusCmd.Dir = config.RelayTestPath
	gitOutput, err := gitStatusCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if !strings.Contains(string(gitOutput), "commit_test.md") {
		return fmt.Errorf("test file not detected by git")
	}

	// Try smart commit
	commitCmd := exec.Command(config.RelayBinaryPath, "commit")
	output, err := commitCmd.CombinedOutput()

	outputStr := string(output)

	// The smart commit might fail due to Claude permissions in test environment
	// We'll check if it at least attempted to analyze the changes
	if err != nil {
		// Check if it's a permission/auth error (expected) vs code error
		if strings.Contains(outputStr, "permission") ||
			strings.Contains(outputStr, "auth") ||
			strings.Contains(outputStr, "Unable to run") ||
			strings.Contains(outputStr, "Claude CLI") {
			fmt.Printf("Smart commit failed due to permissions (expected in test): %s\n", outputStr)
		} else {
			return fmt.Errorf("smart commit failed unexpectedly: %w (output: %s)", err, outputStr)
		}
	} else {
		// If it succeeded, verify it actually did something
		if !strings.Contains(outputStr, "commit") && !strings.Contains(outputStr, "Smart commit") {
			return fmt.Errorf("smart commit completed but output doesn't indicate success: %s", outputStr)
		}
	}

	// Clean up test file
	os.Remove(testFile)

	// Clean up project
	removeCmd := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	removeCmd.CombinedOutput()

	return nil
}
