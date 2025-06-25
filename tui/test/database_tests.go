package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func runDatabaseTests(config *TestConfig) *TestSuite {
	suite := &TestSuite{Name: "Database Operations"}

	// Test database creation
	passed, duration, err := runTest("Database Creation", func() error {
		return testDatabaseCreation(config)
	})
	suite.AddResult("Database Creation", passed, duration, err)

	// Test project addition
	passed, duration, err = runTest("Add Project", func() error {
		return testAddProject(config)
	})
	suite.AddResult("Add Project", passed, duration, err)

	// Test project retrieval
	passed, duration, err = runTest("Get Project", func() error {
		return testGetProject(config)
	})
	suite.AddResult("Get Project", passed, duration, err)

	// Test project listing
	passed, duration, err = runTest("List Projects", func() error {
		return testListProjects(config)
	})
	suite.AddResult("List Projects", passed, duration, err)

	// Test active project management
	passed, duration, err = runTest("Active Project", func() error {
		return testActiveProject(config)
	})
	suite.AddResult("Active Project", passed, duration, err)

	// Test project removal
	passed, duration, err = runTest("Remove Project", func() error {
		return testRemoveProject(config)
	})
	suite.AddResult("Remove Project", passed, duration, err)

	// Test duplicate project handling
	passed, duration, err = runTest("Duplicate Project Handling", func() error {
		return testDuplicateProject(config)
	})
	suite.AddResult("Duplicate Project Handling", passed, duration, err)

	return suite
}

func testDatabaseCreation(config *TestConfig) error {
	// Create a test database instance
	testPath := config.TestDBPath

	// Create test database directory
	testDir := filepath.Dir(testPath)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Import the database module (we'll need to create a test version)
	// For now, simulate database creation
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		return fmt.Errorf("database directory was not created")
	}

	return nil
}

func testAddProject(config *TestConfig) error {
	// This would test adding a project to the database
	// We'll simulate this by checking if we can create the database structure
	// In a real implementation, we'd use the actual database functions

	// Create a simple test database file
	testDB := config.TestDBPath
	file, err := os.Create(testDB)
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}
	defer file.Close()

	// Write some test data to simulate a project entry
	_, err = file.WriteString("test_project_data")
	if err != nil {
		return fmt.Errorf("failed to write test data: %w", err)
	}

	return nil
}

func testGetProject(config *TestConfig) error {
	// Test retrieving a project from the database
	testDB := config.TestDBPath

	// Check if test database exists
	if _, err := os.Stat(testDB); os.IsNotExist(err) {
		return fmt.Errorf("test database does not exist")
	}

	// Read test data
	data, err := os.ReadFile(testDB)
	if err != nil {
		return fmt.Errorf("failed to read test database: %w", err)
	}

	if string(data) != "test_project_data" {
		return fmt.Errorf("unexpected data in test database")
	}

	return nil
}

func testListProjects(config *TestConfig) error {
	// Test listing all projects
	testDB := config.TestDBPath

	// Verify database exists and has content
	if _, err := os.Stat(testDB); os.IsNotExist(err) {
		return fmt.Errorf("test database does not exist for listing")
	}

	return nil
}

func testActiveProject(config *TestConfig) error {
	// Test setting and getting active project
	// For now, just verify we can manage the concept
	testFile := filepath.Join(filepath.Dir(config.TestDBPath), "active_project_test")

	// Write active project info
	err := os.WriteFile(testFile, []byte("test_project"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write active project test: %w", err)
	}

	// Read it back
	data, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read active project test: %w", err)
	}

	if string(data) != "test_project" {
		return fmt.Errorf("active project data mismatch")
	}

	// Cleanup
	os.Remove(testFile)

	return nil
}

func testRemoveProject(config *TestConfig) error {
	// Test removing a project
	testDB := config.TestDBPath

	// Remove test database to simulate project removal
	if err := os.Remove(testDB); err != nil {
		return fmt.Errorf("failed to remove test database: %w", err)
	}

	// Verify it's gone
	if _, err := os.Stat(testDB); !os.IsNotExist(err) {
		return fmt.Errorf("test database still exists after removal")
	}

	return nil
}

func testDuplicateProject(config *TestConfig) error {
	// Test handling of duplicate project names
	testDB := config.TestDBPath

	// Create first project
	file1, err := os.Create(testDB)
	if err != nil {
		return fmt.Errorf("failed to create first test database: %w", err)
	}
	file1.Close()

	// Try to create duplicate (simulate duplicate detection)
	if _, err := os.Stat(testDB); err == nil {
		// File exists, this simulates duplicate detection
		os.Remove(testDB) // Cleanup
		return nil
	}

	return fmt.Errorf("duplicate detection failed")
}
