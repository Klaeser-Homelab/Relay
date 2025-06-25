package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	IssueTracker IssueTrackerConfig `json:"issue_tracker"`
	LLMs         LLMConfig          `json:"llms"`
}

// IssueTrackerConfig contains issue tracker settings
type IssueTrackerConfig struct {
	Provider string       `json:"provider"` // "local", "github", etc.
	GitHub   GitHubConfig `json:"github"`
}

// GitHubConfig contains GitHub-specific settings
type GitHubConfig struct {
	Repository    string `json:"repository"`     // "owner/repo" format
	SyncDirection string `json:"sync_direction"` // "bidirectional", "push", "pull"
	AutoSync      bool   `json:"auto_sync"`      // Enable automatic syncing
	SyncInterval  int    `json:"sync_interval"`  // Minutes between syncs (0 = disabled)
	LastSyncedAt  string `json:"last_synced_at"` // ISO timestamp of last successful sync
}

// LLMConfig contains LLM provider settings
type LLMConfig struct {
	Planning  LLMProviderConfig `json:"planning"`  // Planning provider configuration
	Executing LLMProviderConfig `json:"executing"` // Executing provider configuration
}

// ConfigManager manages application configuration
type ConfigManager struct {
	config   Config
	dataFile string
}

// NewConfigManager creates a new config manager
func NewConfigManager(projectPath string) (*ConfigManager, error) {
	configFile := filepath.Join(projectPath, ".relay", "config.json")

	// Ensure the .relay directory exists
	relayDir := filepath.Dir(configFile)
	if err := os.MkdirAll(relayDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	cm := &ConfigManager{
		dataFile: configFile,
		config:   getDefaultConfig(),
	}

	// Load existing config if it exists
	if err := cm.loadConfig(); err != nil {
		// If config doesn't exist, save default config
		if os.IsNotExist(err) {
			if saveErr := cm.saveConfig(); saveErr != nil {
				return nil, fmt.Errorf("failed to save default config: %w", saveErr)
			}
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return cm, nil
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() Config {
	return Config{
		IssueTracker: IssueTrackerConfig{
			Provider: "github",
			GitHub: GitHubConfig{
				Repository:    "", // Will be auto-detected from git remote
				SyncDirection: "bidirectional",
				AutoSync:      false,
				SyncInterval:  0, // Disabled by default
				LastSyncedAt:  "",
			},
		},
		LLMs: LLMConfig{
			Planning: LLMProviderConfig{
				Type:      "claude-cli", // Use backwards compatible CLI by default
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 4096,
				Options:   make(map[string]string),
			},
			Executing: LLMProviderConfig{
				Type:      "claude-cli", // Use backwards compatible CLI by default
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 4096,
				Options:   make(map[string]string),
			},
		},
	}
}

// loadConfig loads configuration from file
func (cm *ConfigManager) loadConfig() error {
	data, err := os.ReadFile(cm.dataFile)
	if err != nil {
		return err
	}

	// Try to unmarshal with the new format first
	if err := json.Unmarshal(data, &cm.config); err != nil {
		// If that fails, try to load with the old format and migrate
		var oldConfig struct {
			IssueTracker IssueTrackerConfig `json:"issue_tracker"`
			LLMs         struct {
				Planning  string `json:"planning"`
				Executing string `json:"executing"`
			} `json:"llms"`
		}

		if err := json.Unmarshal(data, &oldConfig); err != nil {
			return fmt.Errorf("failed to parse config (tried both new and old formats): %w", err)
		}

		// Migrate old format to new format
		cm.config.IssueTracker = oldConfig.IssueTracker
		cm.config.LLMs = LLMConfig{
			Planning: LLMProviderConfig{
				Type:      migrateLLMType(oldConfig.LLMs.Planning),
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 4096,
				Options:   make(map[string]string),
			},
			Executing: LLMProviderConfig{
				Type:      migrateLLMType(oldConfig.LLMs.Executing),
				Model:     "claude-3-5-sonnet-20241022",
				MaxTokens: 4096,
				Options:   make(map[string]string),
			},
		}

		// Save the migrated config
		return cm.saveConfig()
	}

	return nil
}

// migrateLLMType converts old LLM type strings to new format
func migrateLLMType(oldType string) string {
	switch oldType {
	case "claude":
		return "claude-cli" // Use CLI for backwards compatibility
	case "openai":
		return "openai"
	case "local":
		return "local"
	default:
		return "claude-cli" // Default fallback
	}
}

// saveConfig saves configuration to file
func (cm *ConfigManager) saveConfig() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(cm.dataFile, data, 0644)
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() Config {
	return cm.config
}

// UpdateLLMPlanning updates the planning LLM setting
func (cm *ConfigManager) UpdateLLMPlanning(config LLMProviderConfig) error {
	cm.config.LLMs.Planning = config
	return cm.saveConfig()
}

// UpdateLLMExecuting updates the executing LLM setting
func (cm *ConfigManager) UpdateLLMExecuting(config LLMProviderConfig) error {
	cm.config.LLMs.Executing = config
	return cm.saveConfig()
}

// UpdateLLMPlanningType updates just the planning LLM type
func (cm *ConfigManager) UpdateLLMPlanningType(providerType string) error {
	cm.config.LLMs.Planning.Type = providerType
	return cm.saveConfig()
}

// UpdateLLMExecutingType updates just the executing LLM type
func (cm *ConfigManager) UpdateLLMExecutingType(providerType string) error {
	cm.config.LLMs.Executing.Type = providerType
	return cm.saveConfig()
}

// UpdateIssueTracker updates the issue tracker setting
func (cm *ConfigManager) UpdateIssueTracker(provider string) error {
	cm.config.IssueTracker.Provider = provider
	return cm.saveConfig()
}

// UpdateGitHubRepository updates the GitHub repository setting
func (cm *ConfigManager) UpdateGitHubRepository(repo string) error {
	cm.config.IssueTracker.GitHub.Repository = repo
	return cm.saveConfig()
}

// UpdateGitHubSyncDirection updates the GitHub sync direction
func (cm *ConfigManager) UpdateGitHubSyncDirection(direction string) error {
	cm.config.IssueTracker.GitHub.SyncDirection = direction
	return cm.saveConfig()
}

// UpdateGitHubAutoSync updates the GitHub auto-sync setting
func (cm *ConfigManager) UpdateGitHubAutoSync(enabled bool) error {
	cm.config.IssueTracker.GitHub.AutoSync = enabled
	return cm.saveConfig()
}

// UpdateGitHubLastSyncedAt updates the last synced timestamp
func (cm *ConfigManager) UpdateGitHubLastSyncedAt(timestamp string) error {
	cm.config.IssueTracker.GitHub.LastSyncedAt = timestamp
	return cm.saveConfig()
}

// GetGitHubConfig returns the GitHub configuration
func (cm *ConfigManager) GetGitHubConfig() GitHubConfig {
	return cm.config.IssueTracker.GitHub
}
