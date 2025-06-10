package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "add":
		handleAddProject()
	case "open":
		handleOpenProject()
	case "list":
		handleListProjects()
	case "remove":
		handleRemoveProject()
	case "commit":
		handleSmartCommit()
	case "push":
		handleSmartPush()
	case "commit-push":
		handleSmartCommitPush()
	case "status":
		handleProjectStatus()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Relay Server - AI-powered project management")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  relay add -p <path>     Add a new project")
	fmt.Println("  relay open <name>       Open/switch to a project")
	fmt.Println("  relay list              List all projects")
	fmt.Println("  relay remove <name>     Remove a project")
	fmt.Println("  relay commit            Smart commit with AI-generated message")
	fmt.Println("  relay push              Push to current branch")
	fmt.Println("  relay commit-push       Smart commit and push")
	fmt.Println("  relay status            Show current project status")
}

func handleAddProject() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	projectPath := addCmd.String("p", "", "Project path")
	addCmd.Parse(os.Args[2:])
	
	if *projectPath == "" {
		fmt.Println("Error: Project path is required. Use -p flag.")
		return
	}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(*projectPath)
	if err != nil {
		fmt.Printf("Error: Invalid path: %v\n", err)
		return
	}
	
	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Error: Directory does not exist: %s\n", absPath)
		return
	}
	
	// Extract project name from path
	projectName := filepath.Base(absPath)
	
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	err = pm.AddProject(projectName, absPath)
	if err != nil {
		fmt.Printf("Error adding project: %v\n", err)
		return
	}
	
	fmt.Printf("Successfully added project '%s' at %s\n", projectName, absPath)
}

func handleOpenProject() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Project name is required")
		return
	}
	
	projectName := os.Args[2]
	
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	err = pm.OpenProject(projectName)
	if err != nil {
		fmt.Printf("Error opening project: %v\n", err)
		return
	}
	
	fmt.Printf("Switched to project '%s'\n", projectName)
}

func handleListProjects() {
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	projects, err := pm.ListProjects()
	if err != nil {
		fmt.Printf("Error listing projects: %v\n", err)
		return
	}
	
	if len(projects) == 0 {
		fmt.Println("No projects found. Use 'relay add -p <path>' to add a project.")
		return
	}
	
	currentProject, _ := pm.GetActiveProject()
	
	fmt.Println("Projects:")
	for _, project := range projects {
		marker := "  "
		if currentProject != nil && project.Name == currentProject.Name {
			marker = "* "
		}
		fmt.Printf("%s%s - %s\n", marker, project.Name, project.Path)
	}
}

func handleRemoveProject() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Project name is required")
		return
	}
	
	projectName := os.Args[2]
	
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	err = pm.RemoveProject(projectName)
	if err != nil {
		fmt.Printf("Error removing project: %v\n", err)
		return
	}
	
	fmt.Printf("Successfully removed project '%s'\n", projectName)
}

func handleSmartCommit() {
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	project, err := pm.GetActiveProject()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Use 'relay open <project>' to select a project first")
		return
	}
	
	gitOps, err := NewGitOperations(project.Path)
	if err != nil {
		fmt.Printf("Error initializing git operations: %v\n", err)
		return
	}
	
	err = gitOps.SmartCommit()
	if err != nil {
		fmt.Printf("Error during smart commit: %v\n", err)
		return
	}
	
	fmt.Println("Smart commit completed successfully")
}

func handleSmartPush() {
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	project, err := pm.GetActiveProject()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Use 'relay open <project>' to select a project first")
		return
	}
	
	gitOps, err := NewGitOperations(project.Path)
	if err != nil {
		fmt.Printf("Error initializing git operations: %v\n", err)
		return
	}
	
	err = gitOps.Push("")
	if err != nil {
		fmt.Printf("Error during push: %v\n", err)
		return
	}
	
	fmt.Println("Push completed successfully")
}

func handleSmartCommitPush() {
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	project, err := pm.GetActiveProject()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Use 'relay open <project>' to select a project first")
		return
	}
	
	gitOps, err := NewGitOperations(project.Path)
	if err != nil {
		fmt.Printf("Error initializing git operations: %v\n", err)
		return
	}
	
	err = gitOps.SmartCommitAndPush()
	if err != nil {
		fmt.Printf("Error during smart commit and push: %v\n", err)
		return
	}
	
	fmt.Println("Smart commit and push completed successfully")
}

func handleProjectStatus() {
	pm, err := NewProjectManager()
	if err != nil {
		log.Printf("Failed to initialize project manager: %v", err)
		return
	}
	defer pm.Close()
	
	project, err := pm.GetActiveProject()
	if err != nil {
		fmt.Printf("No active project. Use 'relay open <project>' to select one.\n")
		return
	}
	
	fmt.Printf("Active Project: %s\n", project.Name)
	fmt.Printf("Path: %s\n", project.Path)
	fmt.Printf("Last Opened: %s\n", project.LastOpened.Format("2006-01-02 15:04:05"))
}