package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Test configuration
type TestConfig struct {
	RelayTestPath   string
	TestDBPath      string
	RelayBinaryPath string
	BackupDBPath    string
}

func getTestConfig() *TestConfig {
	homeDir, _ := os.UserHomeDir()
	return &TestConfig{
		RelayTestPath:   "/Users/reed/Code/Personal/RelayTest",
		TestDBPath:      filepath.Join(homeDir, ".relay", "test_relay.db"),
		RelayBinaryPath: "../relay",
		BackupDBPath:    filepath.Join(homeDir, ".relay", "relay.db.backup"),
	}
}

// Test utilities
func setupTestEnvironment(config *TestConfig) error {
	fmt.Println("🔧 Setting up test environment...")

	// Backup existing database if it exists
	homeDir, _ := os.UserHomeDir()
	originalDB := filepath.Join(homeDir, ".relay", "relay.db")
	if _, err := os.Stat(originalDB); err == nil {
		if err := copyFile(originalDB, config.BackupDBPath); err != nil {
			return fmt.Errorf("failed to backup database: %w", err)
		}
		fmt.Println("✅ Backed up existing database")
	}

	// Remove test database if it exists
	if _, err := os.Stat(config.TestDBPath); err == nil {
		os.Remove(config.TestDBPath)
	}

	// Verify RelayTest directory exists
	if _, err := os.Stat(config.RelayTestPath); os.IsNotExist(err) {
		return fmt.Errorf("RelayTest directory not found at %s", config.RelayTestPath)
	}

	fmt.Println("✅ Test environment ready")
	return nil
}

func teardownTestEnvironment(config *TestConfig) error {
	fmt.Println("🧹 Cleaning up test environment...")

	// Remove test database
	if _, err := os.Stat(config.TestDBPath); err == nil {
		os.Remove(config.TestDBPath)
	}

	// Restore original database if backup exists
	homeDir, _ := os.UserHomeDir()
	originalDB := filepath.Join(homeDir, ".relay", "relay.db")
	if _, err := os.Stat(config.BackupDBPath); err == nil {
		if err := copyFile(config.BackupDBPath, originalDB); err != nil {
			return fmt.Errorf("failed to restore database: %w", err)
		}
		os.Remove(config.BackupDBPath)
		fmt.Println("✅ Restored original database")
	}

	fmt.Println("✅ Cleanup complete")
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// Test result tracking
type TestResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Error    string
}

type TestSuite struct {
	Name    string
	Results []TestResult
}

func (ts *TestSuite) AddResult(name string, passed bool, duration time.Duration, err error) {
	result := TestResult{
		Name:     name,
		Passed:   passed,
		Duration: duration,
	}
	if err != nil {
		result.Error = err.Error()
	}
	ts.Results = append(ts.Results, result)
}

func (ts *TestSuite) PrintSummary() {
	fmt.Printf("\n📊 Test Suite: %s\n", ts.Name)
	fmt.Println("----------------------------------------")

	passed := 0
	total := len(ts.Results)
	totalDuration := time.Duration(0)

	for _, result := range ts.Results {
		status := "❌ FAIL"
		if result.Passed {
			status = "✅ PASS"
			passed++
		}

		fmt.Printf("%s %s (%v)\n", status, result.Name, result.Duration)
		if !result.Passed && result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}
		totalDuration += result.Duration
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("Total: %d/%d passed (%v)\n", passed, total, totalDuration)

	if passed == total {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Printf("⚠️  %d test(s) failed\n", total-passed)
	}
}

// Test helper functions
func runTest(name string, testFunc func() error) (bool, time.Duration, error) {
	start := time.Now()
	err := testFunc()
	duration := time.Since(start)

	return err == nil, duration, err
}

func main() {
	// Check if running in CI
	isCI := os.Getenv("CI_MODE") == "true" || os.Getenv("CI") == "true"

	if isCI {
		fmt.Println("🚀 Starting Relay Test Suite (CI Mode)")
	} else {
		fmt.Println("🚀 Starting Relay Test Suite")
	}
	fmt.Println("============================")

	// Use CI-aware configuration
	config := getCITestConfig()

	// Setup test environment (CI-aware)
	if err := setupCITestEnvironment(config); err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	// Defer cleanup (CI-aware)
	defer func() {
		if err := teardownCITestEnvironment(config); err != nil {
			log.Printf("Warning: Failed to cleanup test environment: %v", err)
		}
	}()

	// Run all test suites
	allSuites := []*TestSuite{}

	// Database tests
	dbSuite := runDatabaseTests(config)
	allSuites = append(allSuites, dbSuite)

	// Project management tests
	projectSuite := runProjectTests(config)
	allSuites = append(allSuites, projectSuite)

	// Claude integration tests
	claudeSuite := runClaudeTests(config)
	allSuites = append(allSuites, claudeSuite)

	// Git operations tests
	gitSuite := runGitTests(config)
	allSuites = append(allSuites, gitSuite)

	// End-to-end integration tests
	e2eSuite := runIntegrationTests(config)
	allSuites = append(allSuites, e2eSuite)

	// Print all results
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📋 FINAL RESULTS")
	fmt.Println(strings.Repeat("=", 50))

	totalPassed := 0
	totalTests := 0

	for _, suite := range allSuites {
		suite.PrintSummary()

		for _, result := range suite.Results {
			totalTests++
			if result.Passed {
				totalPassed++
			}
		}
	}

	fmt.Printf("\n🏆 OVERALL: %d/%d tests passed\n", totalPassed, totalTests)

	if totalPassed == totalTests {
		fmt.Println("🎉 All tests passed! Relay is working correctly.")
		os.Exit(0)
	} else {
		fmt.Printf("❌ %d test(s) failed. Please review the failures above.\n", totalTests-totalPassed)
		os.Exit(1)
	}
}
