package main

import (
	"fmt"
	"os"
	"time"
)

type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	LastOpened  time.Time `json:"last_opened"`
	Description string    `json:"description"`
}

type ProjectManager struct {
	db *Database
}

func NewProjectManager() (*ProjectManager, error) {
	db, err := NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &ProjectManager{
		db: db,
	}, nil
}

func (pm *ProjectManager) AddProject(name, path string) error {
	// Validate that path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if project already exists
	_, err := pm.db.GetProject(name)
	if err == nil {
		return fmt.Errorf("project '%s' already exists", name)
	}

	// Add the project
	err = pm.db.AddProject(name, path)
	if err != nil {
		return fmt.Errorf("failed to add project to database: %w", err)
	}

	return nil
}

func (pm *ProjectManager) OpenProject(name string) error {
	// Get the project
	project, err := pm.db.GetProject(name)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Verify the project path still exists
	if _, err := os.Stat(project.Path); os.IsNotExist(err) {
		return fmt.Errorf("project path no longer exists: %s", project.Path)
	}

	// Set as active project
	err = pm.db.SetActiveProject(project.ID)
	if err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}

	// Update last opened timestamp
	err = pm.db.UpdateProjectLastOpened(project.ID)
	if err != nil {
		return fmt.Errorf("failed to update last opened: %w", err)
	}

	// Change working directory to project path
	err = os.Chdir(project.Path)
	if err != nil {
		return fmt.Errorf("failed to change working directory: %w", err)
	}

	return nil
}

func (pm *ProjectManager) ListProjects() ([]*Project, error) {
	projects, err := pm.db.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

func (pm *ProjectManager) RemoveProject(name string) error {
	err := pm.db.RemoveProject(name)
	if err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	return nil
}

func (pm *ProjectManager) GetActiveProject() (*Project, error) {
	project, err := pm.db.GetActiveProject()
	if err != nil {
		return nil, fmt.Errorf("failed to get active project: %w", err)
	}

	return project, nil
}

func (pm *ProjectManager) GetProject(name string) (*Project, error) {
	project, err := pm.db.GetProject(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

func (pm *ProjectManager) EnsureActiveProjectPath() error {
	// Get active project and ensure we're in the right directory
	project, err := pm.GetActiveProject()
	if err != nil {
		return err
	}

	// Check current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// If we're not in the project directory, change to it
	if currentDir != project.Path {
		err = os.Chdir(project.Path)
		if err != nil {
			return fmt.Errorf("failed to change to project directory: %w", err)
		}
	}

	return nil
}

func (pm *ProjectManager) Close() error {
	if pm.db != nil {
		return pm.db.Close()
	}
	return nil
}
