package main

import (
	"fmt"
	"time"
)

// GitHubSyncManager handles bidirectional synchronization between local issues and GitHub
type GitHubSyncManager struct {
	issueManager  *IssueManager
	githubService *GitHubService
	configManager *ConfigManager
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Success        bool     `json:"success"`
	CreatedLocal   int      `json:"created_local"`   // Issues created locally from GitHub
	CreatedGitHub  int      `json:"created_github"`  // Issues created on GitHub from local
	UpdatedLocal   int      `json:"updated_local"`   // Local issues updated from GitHub
	UpdatedGitHub  int      `json:"updated_github"`  // GitHub issues updated from local
	Errors         []string `json:"errors"`          // List of errors encountered
	ConflictsFound int      `json:"conflicts_found"` // Number of conflicts detected
}

// NewGitHubSyncManager creates a new sync manager
func NewGitHubSyncManager(issueManager *IssueManager, githubService *GitHubService, configManager *ConfigManager) *GitHubSyncManager {
	return &GitHubSyncManager{
		issueManager:  issueManager,
		githubService: githubService,
		configManager: configManager,
	}
}

// SyncPull pulls issues from GitHub and creates/updates local issues
func (gsm *GitHubSyncManager) SyncPull() (*SyncResult, error) {
	result := &SyncResult{Success: true}

	// Check GitHub authentication
	authenticated, err := gsm.githubService.IsAuthenticated()
	if err != nil || !authenticated {
		result.Success = false
		result.Errors = append(result.Errors, "GitHub authentication failed")
		return result, fmt.Errorf("GitHub authentication failed")
	}

	// Fetch GitHub issues
	githubIssues, err := gsm.githubService.FetchIssues()
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch GitHub issues: %v", err))
		return result, err
	}

	now := time.Now()

	for _, githubIssue := range githubIssues {
		// Check if we already have this GitHub issue locally
		existingIssue, err := gsm.issueManager.GetIssueByGitHubID(githubIssue.Number)

		if err != nil {
			// Issue doesn't exist locally, create it
			localIssue, err := gsm.createLocalIssueFromGitHub(githubIssue, now)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create local issue from GitHub #%d: %v", githubIssue.Number, err))
				continue
			}
			result.CreatedLocal++

			// Mark as synced
			gsm.issueManager.UpdateIssueSyncStatus(localIssue.ID, "Synced")
		} else {
			// Issue exists locally, check if it needs updating
			if gsm.shouldUpdateLocalIssue(existingIssue, githubIssue) {
				err := gsm.updateLocalIssueFromGitHub(existingIssue, githubIssue, now)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to update local issue %d from GitHub #%d: %v", existingIssue.ID, githubIssue.Number, err))
					continue
				}
				result.UpdatedLocal++
			}
		}
	}

	// Update last synced timestamp in config
	gsm.configManager.UpdateGitHubLastSyncedAt(now.Format(time.RFC3339))

	return result, nil
}

// SyncPush pushes local issues to GitHub
func (gsm *GitHubSyncManager) SyncPush() (*SyncResult, error) {
	result := &SyncResult{Success: true}

	// Check GitHub authentication
	authenticated, err := gsm.githubService.IsAuthenticated()
	if err != nil || !authenticated {
		result.Success = false
		result.Errors = append(result.Errors, "GitHub authentication failed")
		return result, fmt.Errorf("GitHub authentication failed")
	}

	// Get unsynced local issues
	unsyncedIssues := gsm.issueManager.GetUnsyncedIssues()
	now := time.Now()

	for _, localIssue := range unsyncedIssues {
		// Mark as syncing
		gsm.issueManager.UpdateIssueSyncStatus(localIssue.ID, "Syncing")

		if localIssue.GitHubID == nil {
			// Create new GitHub issue
			githubIssue, err := gsm.createGitHubIssueFromLocal(localIssue)
			if err != nil {
				gsm.issueManager.UpdateIssueSyncStatus(localIssue.ID, "Sync Error")
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create GitHub issue from local #%d: %v", localIssue.ID, err))
				continue
			}

			// Update local issue with GitHub data
			err = gsm.issueManager.UpdateIssueGitHubData(localIssue.ID, &githubIssue.Number, githubIssue.URL, &now)
			if err != nil {
				gsm.issueManager.UpdateIssueSyncStatus(localIssue.ID, "Sync Error")
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to update local issue %d with GitHub data: %v", localIssue.ID, err))
				continue
			}

			result.CreatedGitHub++
		} else {
			// Update existing GitHub issue
			err := gsm.updateGitHubIssueFromLocal(localIssue)
			if err != nil {
				gsm.issueManager.UpdateIssueSyncStatus(localIssue.ID, "Sync Error")
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to update GitHub issue #%d from local #%d: %v", *localIssue.GitHubID, localIssue.ID, err))
				continue
			}

			// Update last synced timestamp
			err = gsm.issueManager.UpdateIssueGitHubData(localIssue.ID, localIssue.GitHubID, localIssue.GitHubURL, &now)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to update sync timestamp for issue %d: %v", localIssue.ID, err))
			}

			result.UpdatedGitHub++
		}

		// Mark as synced if no errors
		gsm.issueManager.UpdateIssueSyncStatus(localIssue.ID, "Synced")
	}

	// Update last synced timestamp in config
	gsm.configManager.UpdateGitHubLastSyncedAt(now.Format(time.RFC3339))

	return result, nil
}

// SyncBidirectional performs a full bidirectional sync
func (gsm *GitHubSyncManager) SyncBidirectional() (*SyncResult, error) {
	// First pull from GitHub
	pullResult, err := gsm.SyncPull()
	if err != nil {
		return pullResult, err
	}

	// Then push to GitHub
	pushResult, err := gsm.SyncPush()
	if err != nil {
		// Combine results
		pullResult.Errors = append(pullResult.Errors, pushResult.Errors...)
		pullResult.Success = false
		return pullResult, err
	}

	// Combine results
	combinedResult := &SyncResult{
		Success:        pullResult.Success && pushResult.Success,
		CreatedLocal:   pullResult.CreatedLocal,
		CreatedGitHub:  pushResult.CreatedGitHub,
		UpdatedLocal:   pullResult.UpdatedLocal,
		UpdatedGitHub:  pushResult.UpdatedGitHub,
		ConflictsFound: pullResult.ConflictsFound + pushResult.ConflictsFound,
	}

	// Combine errors
	combinedResult.Errors = append(pullResult.Errors, pushResult.Errors...)

	return combinedResult, nil
}

// createLocalIssueFromGitHub creates a new local issue from a GitHub issue
func (gsm *GitHubSyncManager) createLocalIssueFromGitHub(githubIssue GitHubIssue, syncTime time.Time) (*Issue, error) {
	// Map GitHub data to local format
	status := gsm.githubService.MapGitHubStateToLocal(githubIssue.State)
	labels := gsm.githubService.MapGitHubLabelsToLocal(githubIssue.Labels)

	// Create the issue
	localIssue, err := gsm.issueManager.AddIssue(githubIssue.Title)
	if err != nil {
		return nil, err
	}

	// Update with GitHub-specific data
	err = gsm.issueManager.UpdateIssueStatus(localIssue.ID, status)
	if err != nil {
		return nil, err
	}

	err = gsm.issueManager.UpdateIssueLabels(localIssue.ID, labels)
	if err != nil {
		return nil, err
	}

	// Map GitHub Description/Body to Relay Prompt
	prompt := gsm.extractPromptFromGitHubBody(githubIssue.Body)
	if prompt != "" {
		err = gsm.issueManager.UpdateIssuePrompt(localIssue.ID, prompt)
		if err != nil {
			return nil, err
		}
	}

	err = gsm.issueManager.UpdateIssueGitHubData(localIssue.ID, &githubIssue.Number, githubIssue.URL, &syncTime)
	if err != nil {
		return nil, err
	}

	return localIssue, nil
}

// createGitHubIssueFromLocal creates a new GitHub issue from a local issue
func (gsm *GitHubSyncManager) createGitHubIssueFromLocal(localIssue Issue) (*GitHubIssue, error) {
	// Map local data to GitHub format
	labels := gsm.githubService.MapLocalLabelsToGitHub(localIssue.Labels)

	// Map Relay Prompt to GitHub Description/Body
	body := localIssue.Prompt

	// Create the GitHub issue
	githubIssue, err := gsm.githubService.CreateIssue(localIssue.Content, body, labels)
	if err != nil {
		return nil, err
	}

	return githubIssue, nil
}

// updateLocalIssueFromGitHub updates a local issue with data from GitHub using smart merging
func (gsm *GitHubSyncManager) updateLocalIssueFromGitHub(localIssue *Issue, githubIssue GitHubIssue, syncTime time.Time) error {
	// Map GitHub data to local format
	status := gsm.githubService.MapGitHubStateToLocal(githubIssue.State)
	labels := gsm.githubService.MapGitHubLabelsToLocal(githubIssue.Labels)

	// Extract prompt from GitHub body
	githubPrompt := gsm.extractPromptFromGitHubBody(githubIssue.Body)

	// Smart field-level merging with newest-wins for conflicts
	lastSyncTime := localIssue.LastSyncedAt
	if lastSyncTime == nil {
		// No previous sync, GitHub wins for all fields
		if localIssue.Content != githubIssue.Title {
			err := gsm.issueManager.UpdateIssueContent(localIssue.ID, githubIssue.Title)
			if err != nil {
				return err
			}
		}
		if localIssue.Status != status {
			err := gsm.issueManager.UpdateIssueStatus(localIssue.ID, status)
			if err != nil {
				return err
			}
		}
		if !equalSlices(localIssue.Labels, labels) {
			err := gsm.issueManager.UpdateIssueLabels(localIssue.ID, labels)
			if err != nil {
				return err
			}
		}
		if localIssue.Prompt != githubPrompt {
			err := gsm.issueManager.UpdateIssuePrompt(localIssue.ID, githubPrompt)
			if err != nil {
				return err
			}
		}
	} else {
		// Check each field for conflicts and apply newest-wins strategy

		// Content/Title - newest-wins if both changed since last sync
		if localIssue.Content != githubIssue.Title {
			localChangedAfterSync := localIssue.Timestamp.After(*lastSyncTime)
			githubChangedAfterSync := githubIssue.UpdatedAt.After(*lastSyncTime)

			if githubChangedAfterSync && (!localChangedAfterSync || githubIssue.UpdatedAt.After(localIssue.Timestamp)) {
				err := gsm.issueManager.UpdateIssueContent(localIssue.ID, githubIssue.Title)
				if err != nil {
					return err
				}
			}
		}

		// Status - newest-wins if both changed since last sync
		if localIssue.Status != status {
			localChangedAfterSync := localIssue.Timestamp.After(*lastSyncTime)
			githubChangedAfterSync := githubIssue.UpdatedAt.After(*lastSyncTime)

			if githubChangedAfterSync && (!localChangedAfterSync || githubIssue.UpdatedAt.After(localIssue.Timestamp)) {
				err := gsm.issueManager.UpdateIssueStatus(localIssue.ID, status)
				if err != nil {
					return err
				}
			}
		}

		// Labels - merge automatically if no conflict, newest-wins for conflicts
		if !equalSlices(localIssue.Labels, labels) {
			localChangedAfterSync := localIssue.Timestamp.After(*lastSyncTime)
			githubChangedAfterSync := githubIssue.UpdatedAt.After(*lastSyncTime)

			if githubChangedAfterSync && (!localChangedAfterSync || githubIssue.UpdatedAt.After(localIssue.Timestamp)) {
				err := gsm.issueManager.UpdateIssueLabels(localIssue.ID, labels)
				if err != nil {
					return err
				}
			}
		}

		// Prompt - newest-wins if both changed since last sync
		if localIssue.Prompt != githubPrompt {
			localChangedAfterSync := localIssue.Timestamp.After(*lastSyncTime)
			githubChangedAfterSync := githubIssue.UpdatedAt.After(*lastSyncTime)

			if githubChangedAfterSync && (!localChangedAfterSync || githubIssue.UpdatedAt.After(localIssue.Timestamp)) {
				err := gsm.issueManager.UpdateIssuePrompt(localIssue.ID, githubPrompt)
				if err != nil {
					return err
				}
			}
		}
	}

	// Update GitHub metadata
	err := gsm.issueManager.UpdateIssueGitHubData(localIssue.ID, &githubIssue.Number, githubIssue.URL, &syncTime)
	if err != nil {
		return err
	}

	return nil
}

// updateGitHubIssueFromLocal updates a GitHub issue with data from local issue
func (gsm *GitHubSyncManager) updateGitHubIssueFromLocal(localIssue Issue) error {
	if localIssue.GitHubID == nil {
		return fmt.Errorf("local issue %d has no GitHub ID", localIssue.ID)
	}

	// Map local data to GitHub format
	githubState := gsm.githubService.MapLocalStatusToGitHub(localIssue.Status)
	labels := gsm.githubService.MapLocalLabelsToGitHub(localIssue.Labels)

	// Map Relay Prompt directly to GitHub Description/Body
	body := localIssue.Prompt

	// Update the GitHub issue
	err := gsm.githubService.UpdateIssue(*localIssue.GitHubID, localIssue.Content, body, githubState, labels)
	if err != nil {
		return err
	}

	return nil
}

// shouldUpdateLocalIssue determines if a local issue should be updated from GitHub
func (gsm *GitHubSyncManager) shouldUpdateLocalIssue(localIssue *Issue, githubIssue GitHubIssue) bool {
	// Always update if no previous sync
	if localIssue.LastSyncedAt == nil {
		return true
	}

	// Check if GitHub issue was updated after our last sync
	if githubIssue.UpdatedAt.Before(*localIssue.LastSyncedAt) {
		return false // GitHub issue is older than our last sync
	}

	// Check if any field is different (we'll handle conflict resolution in the update function)
	expectedStatus := gsm.githubService.MapGitHubStateToLocal(githubIssue.State)
	expectedLabels := gsm.githubService.MapGitHubLabelsToLocal(githubIssue.Labels)
	expectedPrompt := gsm.extractPromptFromGitHubBody(githubIssue.Body)

	return localIssue.Content != githubIssue.Title ||
		localIssue.Status != expectedStatus ||
		!equalSlices(localIssue.Labels, expectedLabels) ||
		localIssue.Prompt != expectedPrompt
}

// equalSlices compares two string slices for equality
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// extractPromptFromGitHubBody extracts the prompt from GitHub issue body
// For now, we simply treat the entire body as the prompt since we're mapping Prompt <-> Body directly
func (gsm *GitHubSyncManager) extractPromptFromGitHubBody(body string) string {
	// Direct mapping: GitHub body becomes Relay prompt
	return body
}
