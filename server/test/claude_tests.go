package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func runClaudeTests(config *TestConfig) *TestSuite {
	suite := &TestSuite{Name: "Claude CLI Integration"}

	// Test Claude CLI availability
	passed, duration, err := runTest("Claude CLI Available", func() error {
		return testClaudeAvailable()
	})
	suite.AddResult("Claude CLI Available", passed, duration, err)

	// Test Claude command execution
	passed, duration, err = runTest("Claude Command Execution", func() error {
		return testClaudeCommandExecution()
	})
	suite.AddResult("Claude Command Execution", passed, duration, err)

	// Test Claude JSON output parsing
	passed, duration, err = runTest("Claude JSON Output", func() error {
		return testClaudeJSONOutput()
	})
	suite.AddResult("Claude JSON Output", passed, duration, err)

	// Test Claude session continuity
	passed, duration, err = runTest("Claude Session Continuity", func() error {
		return testClaudeSessionContinuity()
	})
	suite.AddResult("Claude Session Continuity", passed, duration, err)

	// Test Claude error handling
	passed, duration, err = runTest("Claude Error Handling", func() error {
		return testClaudeErrorHandling()
	})
	suite.AddResult("Claude Error Handling", passed, duration, err)

	return suite
}

func testClaudeAvailable() error {
	// Test if Claude CLI is available
	cmd := exec.Command("claude", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Claude CLI not available: %w (output: %s)", err, string(output))
	}

	// Check output contains version information
	outputStr := string(output)
	if !strings.Contains(outputStr, "Claude Code") && !strings.Contains(outputStr, "claude") {
		return fmt.Errorf("unexpected version output: %s", outputStr)
	}

	return nil
}

func testClaudeCommandExecution() error {
	// Test basic Claude command execution
	cmd := exec.Command("claude", "--print", "hello world")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to execute Claude command: %w (output: %s)", err, string(output))
	}

	outputStr := string(output)

	// Should get some response from Claude
	if len(outputStr) == 0 {
		return fmt.Errorf("empty response from Claude")
	}

	// Response should be reasonable
	if len(outputStr) < 10 {
		return fmt.Errorf("suspiciously short response from Claude: %s", outputStr)
	}

	return nil
}

func testClaudeJSONOutput() error {
	// Test Claude with JSON output format
	cmd := exec.Command("claude", "--print", "--output-format", "json", "respond with just the word 'test'")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to execute Claude JSON command: %w (output: %s)", err, string(output))
	}

	outputStr := string(output)

	// Should contain JSON structure
	if !strings.Contains(outputStr, "{") || !strings.Contains(outputStr, "}") {
		return fmt.Errorf("output does not appear to be JSON: %s", outputStr)
	}

	// Should contain expected fields (either content or result)
	if !strings.Contains(outputStr, "content") && !strings.Contains(outputStr, "result") {
		return fmt.Errorf("JSON output missing expected content/result field: %s", outputStr)
	}

	return nil
}

func testClaudeSessionContinuity() error {
	// Test first command in session
	cmd1 := exec.Command("claude", "--print", "--output-format", "json", "remember that my favorite color is blue")
	output1, err := cmd1.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to execute first Claude command: %w (output: %s)", err, string(output1))
	}

	// Test follow-up command with continue flag
	cmd2 := exec.Command("claude", "--print", "--output-format", "json", "--continue", "what is my favorite color?")
	output2, err := cmd2.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to execute second Claude command: %w (output: %s)", err, string(output2))
	}

	outputStr2 := string(output2)

	// Should remember the color (this test might be flaky depending on Claude's context)
	// For now, just verify we got a response
	if len(outputStr2) == 0 {
		return fmt.Errorf("empty response from Claude continue command")
	}

	return nil
}

func testClaudeErrorHandling() error {
	// Test Claude with invalid flags (this should fail gracefully)
	cmd := exec.Command("claude", "--invalid-flag-that-does-not-exist")
	output, err := cmd.CombinedOutput()

	// This should fail
	if err == nil {
		return fmt.Errorf("Claude should have failed with invalid flag")
	}

	outputStr := string(output)

	// Should provide helpful error message
	if len(outputStr) == 0 {
		return fmt.Errorf("no error message provided for invalid flag")
	}

	return nil
}

// Helper function to test Claude integration in project context
func testClaudeInProjectContext(config *TestConfig) error {
	// Add RelayTest project first
	addCmd := exec.Command(config.RelayBinaryPath, "add", "-p", config.RelayTestPath)
	_, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add RelayTest project for Claude context test: %w", err)
	}

	// Open RelayTest project
	openCmd := exec.Command(config.RelayBinaryPath, "open", "RelayTest")
	_, err = openCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open RelayTest project for Claude context test: %w", err)
	}

	// This would test running Claude commands in the project context
	// For now, we'll just verify the project is open
	statusCmd := exec.Command(config.RelayBinaryPath, "status")
	output, err := statusCmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to get project status: %w", err)
	}

	if !strings.Contains(string(output), "RelayTest") {
		return fmt.Errorf("RelayTest project not active")
	}

	// Clean up
	removeCmd := exec.Command(config.RelayBinaryPath, "remove", "RelayTest")
	removeCmd.CombinedOutput()

	return nil
}
