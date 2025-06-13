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

func (g *GitOperations) ListBranches() (string, error) {
	g.logger.Println("Listing git branches...")

	command := "List all git branches (local and remote) and show which one is currently active."

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return "", fmt.Errorf("failed to list branches via Claude: %w", err)
	}

	return response, nil
}

func (g *GitOperations) DeleteBranch(branchName string, force bool) error {
	g.logger.Printf("Deleting local branch: %s (force: %v)", branchName, force)

	var command string
	if force {
		command = fmt.Sprintf("Force delete the local git branch '%s' using 'git branch -D %s'", branchName, branchName)
	} else {
		command = fmt.Sprintf("Delete the local git branch '%s' using 'git branch -d %s'", branchName, branchName)
	}

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to delete branch via Claude: %w", err)
	}

	g.logger.Printf("Delete branch response: %s", response)
	return nil
}

func (g *GitOperations) DeleteRemoteBranch(branchName string) error {
	g.logger.Printf("Deleting remote branch: %s", branchName)

	command := fmt.Sprintf("Delete the remote git branch '%s' using 'git push origin --delete %s'", branchName, branchName)

	response, err := g.llmProvider.SendMessage(context.Background(), command)
	if err != nil {
		return fmt.Errorf("failed to delete remote branch via Claude: %w", err)
	}

	g.logger.Printf("Delete remote branch response: %s", response)
	return nil
}

func (g *GitOperations) Close() error {
	if g.llmProvider != nil {
		return g.llmProvider.Close()
	}
	return nil
}