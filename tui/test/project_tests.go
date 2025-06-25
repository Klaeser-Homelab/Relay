package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func runProjectTests(config *TestConfig) *TestSuite {
	suite := &TestSuite{Name: "Project Management"}

	// Test adding RelayTest project
	passed, duration, err := runTest("Add RelayTest Project", func() error {
		return testAddRelayTestProject(config)
	})
	suite.AddResult("Add RelayTest Project", passed, duration, err)

	// Test listing projects with relay binary
	passed, duration, err = runTest("List Projects CLI", func() error {
		return testListProjectsCLI(config)
	})
	suite.AddResult("List Projects CLI", passed, duration, err)

	// Test opening project
	passed, duration, err = runTest("Open Project", func() error {
		return testOpenProject(config)
	})
	suite.AddResult("Open Project", passed, duration, err)

	// Test project status
	passed, duration, err = runTest("Project Status", func() error {
		return testProjectStatus(config)
	})
	suite.AddResult("Project Status", passed, duration, err)

	// Test invalid project handling
	passed, duration, err = runTest("Invalid Project Handling", func() error {
		return testInvalidProject(config)
	})
	suite.AddResult("Invalid Project Handling", passed, duration, err)

	// Test project removal
	passed, duration, err = runTest("Remove Project CLI", func() error {
		return testRemoveProjectCLI(config)
	})
	suite.AddResult("Remove Project CLI", passed, duration, err)

	return suite
}

func testAddRelayTestProject(config *TestConfig) error {
	// Use the actual relay binary to add RelayTest project
	cmd := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to add RelayTest project: %w (output: %s)", err, string(output))
	}

	// Check output contains success message
	outputStr := string(output)
	if !strings.Contains(outputStr, "Successfully added project") {
		return fmt.Errorf("unexpected output when adding project: %s", outputStr)
	}

	if !strings.Contains(outputStr, "RelayTest") {
		return fmt.Errorf("output does not mention RelayTest project: %s", outputStr)
	}

	return nil
}

func testListProjectsCLI(config *TestConfig) error {
	// List projects using relay binary
	cmd := exec.Command(config.RelayBinaryPath, "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to list projects: %w (output: %s)", err, string(output))
	}

	outputStr := string(output)

	// Should show RelayTest project
	if !strings.Contains(outputStr, "RelayTest") {
		return fmt.Errorf("RelayTest project not found in list output: %s", outputStr)
	}

	// Should show the path
	if !strings.Contains(outputStr, config.RelayTestPath) {
		return fmt.Errorf("RelayTest path not found in list output: %s", outputStr)
	}

	return nil
}

func testOpenProject(config *TestConfig) error {
	// Open RelayTest project
	cmd := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to open RelayTest project: %w (output: %s)", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Switched to project 'RelayTest'") {
		return fmt.Errorf("unexpected output when opening project: %s", outputStr)
	}

	return nil
}

func testProjectStatus(config *TestConfig) error {
	// Check project status
	cmd := exec.Command(config.RelayBinaryPath, "status")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to get project status: %w (output: %s)", err, string(output))
	}

	outputStr := string(output)

	// Should show RelayTest as active project
	if !strings.Contains(outputStr, "Active Project: RelayTest") {
		return fmt.Errorf("RelayTest not shown as active project: %s", outputStr)
	}

	// Should show the correct path
	if !strings.Contains(outputStr, config.RelayTestPath) {
		return fmt.Errorf("RelayTest path not shown in status: %s", outputStr)
	}

	return nil
}

func testInvalidProject(config *TestConfig) error {
	// Try to open a non-existent project
	cmd := exec.Command(config.RelayBinaryPath, "open", "NonExistentProject")
	output, err := cmd.CombinedOutput()

	// This should fail
	if err == nil {
		return fmt.Errorf("opening non-existent project should have failed")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error") {
		return fmt.Errorf("error message not shown for non-existent project: %s", outputStr)
	}

	return nil
}

func testRemoveProjectCLI(config *TestConfig) error {
	// Remove RelayTest project
	cmd := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to remove RelayTest project: %w (output: %s)", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Successfully removed project 'RelayTest'") {
		return fmt.Errorf("unexpected output when removing project: %s", outputStr)
	}

	// Verify project is no longer listed
	cmd = exec.Command(config.RelayBinaryPath, "list")
	output, err = cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to list projects after removal: %w", err)
	}

	outputStr = string(output)

	// Should not contain RelayTest anymore
	if strings.Contains(outputStr, "RelayTest") {
		return fmt.Errorf("RelayTest project still appears in list after removal: %s", outputStr)
	}

	return nil
}
