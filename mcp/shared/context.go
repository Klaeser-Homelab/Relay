package shared

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ExtractIssueNumberFromBranch extracts issue number from branch name
// Supports patterns like: feature/issue-16, bugfix/issue-42, etc.
func ExtractIssueNumberFromBranch(branchName string) (int, error) {
	// Match pattern: any-prefix/issue-{number}
	re := regexp.MustCompile(`(?i)[\w-]+/issue-(\d+)`)
	matches := re.FindStringSubmatch(branchName)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no issue number found in branch name: %s", branchName)
	}
	return strconv.Atoi(matches[1])
}

// GetCurrentBranch gets the current git branch name
func GetCurrentBranch(workingDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workingDir
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// ExtractIssueNumbersFromText extracts issue numbers from text content
// Supports patterns like: #16, Issue #16, Issue: 16, etc.
func ExtractIssueNumbersFromText(text string) []int {
	var numbers []int
	
	// Pattern matches: #16, Issue #16, Issue: 16, issue 16, etc.
	re := regexp.MustCompile(`(?i)(?:#(\d+)|issue[:\s#]+(\d+))`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	for _, match := range matches {
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				if num, err := strconv.Atoi(match[i]); err == nil {
					numbers = append(numbers, num)
				}
			}
		}
	}
	
	return numbers
}

// ExtractPlanFromContext extracts plan content from conversation context
func ExtractPlanFromContext(context string) string {
	// Look for plan sections in the context
	lines := strings.Split(context, "\n")
	var planLines []string
	inPlan := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Start of plan section
		if strings.Contains(strings.ToLower(line), "plan") && 
		   (strings.Contains(line, ":") || strings.Contains(line, "for")) {
			inPlan = true
			planLines = append(planLines, line)
			continue
		}
		
		// If we're in a plan section
		if inPlan {
			// End of plan section (empty line or new section)
			if line == "" || 
			   (strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "##")) {
				break
			}
			planLines = append(planLines, line)
		}
	}
	
	return strings.Join(planLines, "\n")
}

// GitAdd stages all changes in the working directory
func GitAdd(workingDir string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage changes: %s", string(output))
	}
	
	return nil
}

// GitCommit commits staged changes with the given message
func GitCommit(workingDir, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %s", string(output))
	}
	
	return nil
}

// GitPush pushes the current branch to origin
func GitPush(workingDir string) error {
	cmd := exec.Command("git", "push", "origin", "HEAD")
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push changes: %s", string(output))
	}
	
	return nil
}

// GitCheckout switches to the specified branch
func GitCheckout(workingDir, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %s", branchName, string(output))
	}
	
	return nil
}

// DeleteWorktree removes a git worktree directory
func DeleteWorktree(mainRepoDir, worktreeDir string) error {
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreeDir)
	cmd.Dir = mainRepoDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree %s: %s", worktreeDir, string(output))
	}
	
	return nil
}

// DeleteBranch removes a git branch both locally and remotely
func DeleteBranch(workingDir, branchName string) error {
	// Delete local branch
	cmd := exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete local branch %s: %s", branchName, string(output))
	}
	
	// Delete remote branch if it exists
	cmd = exec.Command("git", "push", "origin", "--delete", branchName)
	cmd.Dir = workingDir
	
	output, err = cmd.CombinedOutput()
	if err != nil {
		// Don't fail if remote branch doesn't exist
		if !strings.Contains(string(output), "remote ref does not exist") {
			return fmt.Errorf("failed to delete remote branch %s: %s", branchName, string(output))
		}
	}
	
	return nil
}