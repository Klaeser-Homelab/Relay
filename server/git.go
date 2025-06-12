package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

type GitOperations struct {
	projectPath string
	llmProvider LLMProvider
	logger      *log.Logger
}

func NewGitOperations(projectPath string, llmProvider LLMProvider) (*GitOperations, error) {
	logger := log.New(os.Stdout, "[GitOps] ", log.LstdFlags)

	return &GitOperations{
		projectPath: projectPath,
		llmProvider: llmProvider,
		logger:      logger,
	}, nil
}

func (g *GitOperations) SmartCommit() error {
	g.logger.Println("Starting smart commit process...")

	// Use Claude to analyze changes and create a commit
	command := `Analyze the current git changes in this repository and create an appropriate commit. 
	
	Please:
	1. Run 'git status' to see what files have changed
	2. Run 'git diff' to understand the nature of the changes
	3. Read any changed files if needed to understand the context
	4. Generate a clear, descriptive commit message following conventional commit format when appropriate
	5. Stage all changes and commit with your generated message
	6. Provide a summary of what was committed
	
	Do not ask for confirmation - proceed with the commit.`

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to execute smart commit via Claude: %w", err)
	}

	g.logger.Printf("Smart commit response: %s", response)
	fmt.Printf("Smart commit result:\n%s\n", response)

	return nil
}

func (g *GitOperations) Push(branch string) error {
	g.logger.Printf("Starting push to branch: %s", branch)

	var command string
	if branch == "" {
		command = "Push the current branch to the remote repository. If no upstream is set, set it automatically."
	} else {
		command = fmt.Sprintf("Push the current branch to the remote repository on branch '%s'.", branch)
	}

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to execute push via Claude: %w", err)
	}

	g.logger.Printf("Push response: %s", response)
	fmt.Printf("Push result:\n%s\n", response)

	return nil
}

func (g *GitOperations) SmartCommitAndPush() error {
	g.logger.Println("Starting smart commit and push process...")

	// Use Claude to analyze, commit, and push in one operation
	command := `Analyze the current git changes, create an appropriate commit, and push to the remote repository.
	
	Please:
	1. Run 'git status' to see what files have changed
	2. Run 'git diff' to understand the nature of the changes
	3. Read any changed files if needed to understand the context
	4. Generate a clear, descriptive commit message following conventional commit format when appropriate
	5. Stage all changes and commit with your generated message
	6. Push to the remote repository (set upstream if needed)
	7. Provide a summary of what was committed and pushed
	
	Do not ask for confirmation - proceed with the commit and push.`

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to execute smart commit and push via Claude: %w", err)
	}

	g.logger.Printf("Smart commit and push response: %s", response)
	fmt.Printf("Smart commit and push result:\n%s\n", response)

	return nil
}

func (g *GitOperations) AnalyzeChanges() (string, error) {
	g.logger.Println("Analyzing git changes...")

	command := `Analyze the current git changes in this repository and provide a summary.
	
	Please:
	1. Run 'git status' to see what files have changed
	2. Run 'git diff' to understand the nature of the changes
	3. Provide a summary of the changes without committing anything
	4. Suggest what an appropriate commit message would be
	
	Do not make any commits - just analyze and report.`

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return "", fmt.Errorf("failed to analyze changes via Claude: %w", err)
	}

	g.logger.Printf("Change analysis response: %s", response)

	return response, nil
}

func (g *GitOperations) Status() (string, error) {
	g.logger.Println("Getting git status...")

	command := "Show me the current git status and a brief summary of any changes."

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return "", fmt.Errorf("failed to get git status via Claude: %w", err)
	}

	return response, nil
}

func (g *GitOperations) CreateBranch(branchName string) error {
	g.logger.Printf("Creating branch: %s", branchName)

	command := fmt.Sprintf("Create a new git branch called '%s' and switch to it.", branchName)

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to create branch via Claude: %w", err)
	}

	g.logger.Printf("Create branch response: %s", response)
	fmt.Printf("Create branch result:\n%s\n", response)

	return nil
}

func (g *GitOperations) SwitchBranch(branchName string) error {
	g.logger.Printf("Switching to branch: %s", branchName)

	command := fmt.Sprintf("Switch to git branch '%s'.", branchName)

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to switch branch via Claude: %w", err)
	}

	g.logger.Printf("Switch branch response: %s", response)
	fmt.Printf("Switch branch result:\n%s\n", response)

	return nil
}

func (g *GitOperations) ListBranches() (string, error) {
	g.logger.Println("Listing git branches...")

	command := "List all git branches (local and remote) and show which one is currently active."

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return "", fmt.Errorf("failed to list branches via Claude: %w", err)
	}

	return response, nil
}

func (g *GitOperations) ShowLog(limit int) (string, error) {
	g.logger.Printf("Getting git log (limit: %d)...", limit)

	var command string
	if limit > 0 {
		command = fmt.Sprintf("Show the last %d git commits with their messages and authors.", limit)
	} else {
		command = "Show recent git commits with their messages and authors."
	}

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return "", fmt.Errorf("failed to get git log via Claude: %w", err)
	}

	return response, nil
}

func (g *GitOperations) UndoLastCommit() error {
	g.logger.Println("Undoing last commit...")

	command := "Undo the last git commit while keeping the changes in the working directory (soft reset)."

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to undo last commit via Claude: %w", err)
	}

	g.logger.Printf("Undo commit response: %s", response)
	fmt.Printf("Undo commit result:\n%s\n", response)

	return nil
}

func (g *GitOperations) Close() error {
	if g.llmProvider != nil {
		return g.llmProvider.Close()
	}
	return nil
}

// Helper function to check if directory is a git repository
func (g *GitOperations) IsGitRepository() bool {
	_, err := os.Stat(fmt.Sprintf("%s/.git", g.projectPath))
	return err == nil
}

// Helper function to validate git repository before operations
func (g *GitOperations) validateGitRepo() error {
	if !g.IsGitRepository() {
		return fmt.Errorf("directory is not a git repository: %s", g.projectPath)
	}
	return nil
}

// Enhanced smart commit with validation
func (g *GitOperations) SmartCommitWithValidation() error {
	if err := g.validateGitRepo(); err != nil {
		return err
	}

	return g.SmartCommit()
}

// Enhanced push with validation
func (g *GitOperations) PushWithValidation(branch string) error {
	if err := g.validateGitRepo(); err != nil {
		return err
	}

	return g.Push(branch)
}

// Enhanced smart commit and push with validation
func (g *GitOperations) SmartCommitAndPushWithValidation() error {
	if err := g.validateGitRepo(); err != nil {
		return err
	}

	return g.SmartCommitAndPush()
}
