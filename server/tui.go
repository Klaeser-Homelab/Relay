package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewType represents different screens in the TUI
type ViewType int

const (
	ViewREPL ViewType = iota
	ViewIssueList
	ViewIssueDetail
	ViewTextInput
	ViewConfirmation
	ViewConfig
	ViewLLMConfig
	ViewIssueTrackerConfig
	ViewLabelEditor
)

// Main TUI model that orchestrates different views
type TUIModel struct {
	currentView ViewType
	width       int
	height      int
	
	// REPL components
	replSession  *REPLSession
	replModel    REPLModel
	
	// Issue management components
	issueListModel    IssueListModel
	issueDetailModel  IssueDetailModel
	textInputModel    TextInputModel
	confirmationModel ConfirmationModel
	labelEditorModel  LabelEditorModel
	
	// Config components
	configMenuModel        ConfigMenuModel
	llmConfigModel         LLMConfigModel
	issueTrackerConfigModel IssueTrackerConfigModel
	
	// Navigation state
	previousView ViewType
	err          error
}

// Initialize the TUI
func InitTUI(replSession *REPLSession) TUIModel {
	// Initialize with default dimensions that will be updated by WindowSizeMsg
	defaultWidth, defaultHeight := 80, 24
	
	replModel := NewREPLModel(replSession)
	replModel.width = defaultWidth
	replModel.height = defaultHeight
	
	issueListModel := NewIssueListModel(replSession.issueManager, replSession.configManager, replSession.currentProject.Name, replSession.currentProject.Path)
	issueListModel.width = defaultWidth
	issueListModel.height = defaultHeight
	
	configMenuModel := NewConfigMenuModel(replSession.configManager)
	configMenuModel.width = defaultWidth
	configMenuModel.height = defaultHeight
	
	llmConfigModel := NewLLMConfigModel(replSession.configManager)
	llmConfigModel.width = defaultWidth
	llmConfigModel.height = defaultHeight
	
	issueTrackerConfigModel := NewIssueTrackerConfigModel(replSession.configManager)
	issueTrackerConfigModel.width = defaultWidth
	issueTrackerConfigModel.height = defaultHeight
	
	return TUIModel{
		currentView:             ViewREPL,
		width:                  defaultWidth,
		height:                 defaultHeight,
		replSession:            replSession,
		replModel:              replModel,
		issueListModel:         issueListModel,
		configMenuModel:        configMenuModel,
		llmConfigModel:         llmConfigModel,
		issueTrackerConfigModel: issueTrackerConfigModel,
		previousView:           ViewREPL,
	}
}

func (m TUIModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update all sub-models with new size
		m.replModel.width = msg.Width
		m.replModel.height = msg.Height
		m.issueListModel.width = msg.Width
		m.issueListModel.height = msg.Height
		m.configMenuModel.width = msg.Width
		m.configMenuModel.height = msg.Height
		m.llmConfigModel.width = msg.Width
		m.llmConfigModel.height = msg.Height
		m.issueTrackerConfigModel.width = msg.Width
		m.issueTrackerConfigModel.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
		
		// Handle view-specific navigation
		switch m.currentView {
		case ViewREPL:
			// REPL handles its own commands
		}
	
	case SwitchViewMsg:
		m.previousView = m.currentView
		m.currentView = msg.View
		
		switch msg.View {
		case ViewIssueList:
			m.issueListModel = NewIssueListModel(m.replSession.issueManager, m.replSession.configManager, m.replSession.currentProject.Name, m.replSession.currentProject.Path)
			m.issueListModel.width = m.width
			m.issueListModel.height = m.height
		case ViewIssueDetail:
			if msg.Data != nil {
				if issue, ok := msg.Data.(Issue); ok {
					m.issueDetailModel = NewIssueDetailModel(issue, m.replSession)
				}
			}
		case ViewTextInput:
			if msg.Data != nil {
				if inputData, ok := msg.Data.(TextInputData); ok {
					m.textInputModel = NewTextInputModel(inputData.Prompt, inputData.Placeholder, inputData.OnComplete)
				}
			}
		case ViewConfirmation:
			if msg.Data != nil {
				if confirmData, ok := msg.Data.(ConfirmationData); ok {
					m.confirmationModel = NewConfirmationModel(confirmData.Message, confirmData.OnConfirm)
				}
			}
		case ViewConfig:
			m.configMenuModel = NewConfigMenuModel(m.replSession.configManager)
			m.configMenuModel.width = m.width
			m.configMenuModel.height = m.height
		case ViewLLMConfig:
			m.llmConfigModel = NewLLMConfigModel(m.replSession.configManager)
			m.llmConfigModel.width = m.width
			m.llmConfigModel.height = m.height
		case ViewIssueTrackerConfig:
			m.issueTrackerConfigModel = NewIssueTrackerConfigModel(m.replSession.configManager)
			m.issueTrackerConfigModel.width = m.width
			m.issueTrackerConfigModel.height = m.height
		case ViewLabelEditor:
			if msg.Data != nil {
				if labelData, ok := msg.Data.(LabelEditorData); ok {
					m.labelEditorModel = NewLabelEditorModel(labelData)
					m.labelEditorModel.width = m.width
					m.labelEditorModel.height = m.height
				}
			}
		case ViewREPL:
			// Return to REPL, set context if provided
			if msg.Data != nil {
				if context, ok := msg.Data.(*REPLContext); ok {
					m.replModel.SetContext(context)
				}
			} else {
				// Clear context when returning to REPL without data
				m.replModel.ClearContext()
			}
		}
		
		return m, nil
		
	case BackToPreviousViewMsg:
		m.currentView = m.previousView
		return m, nil
	}
	
	// Update the current view's model
	switch m.currentView {
	case ViewREPL:
		m.replModel, cmd = m.replModel.Update(msg)
	case ViewIssueList:
		m.issueListModel, cmd = m.issueListModel.Update(msg)
	case ViewIssueDetail:
		m.issueDetailModel, cmd = m.issueDetailModel.Update(msg)
	case ViewTextInput:
		m.textInputModel, cmd = m.textInputModel.Update(msg)
	case ViewConfirmation:
		m.confirmationModel, cmd = m.confirmationModel.Update(msg)
	case ViewConfig:
		m.configMenuModel, cmd = m.configMenuModel.Update(msg)
	case ViewLLMConfig:
		m.llmConfigModel, cmd = m.llmConfigModel.Update(msg)
	case ViewIssueTrackerConfig:
		m.issueTrackerConfigModel, cmd = m.issueTrackerConfigModel.Update(msg)
	case ViewLabelEditor:
		m.labelEditorModel, cmd = m.labelEditorModel.Update(msg)
	}
	
	return m, cmd
}

func (m TUIModel) View() string {
	switch m.currentView {
	case ViewREPL:
		return m.replModel.View()
	case ViewIssueList:
		return m.issueListModel.View()
	case ViewIssueDetail:
		return m.issueDetailModel.View()
	case ViewTextInput:
		return m.textInputModel.View()
	case ViewConfirmation:
		return m.confirmationModel.View()
	case ViewConfig:
		return m.configMenuModel.View()
	case ViewLLMConfig:
		return m.llmConfigModel.View()
	case ViewIssueTrackerConfig:
		return m.issueTrackerConfigModel.View()
	case ViewLabelEditor:
		return m.labelEditorModel.View()
	}
	
	return "Unknown view"
}

// Custom messages for view switching and communication
type SwitchViewMsg struct {
	View ViewType
	Data interface{}
}

type BackToPreviousViewMsg struct{}

// Data structures for passing information between views
type TextInputData struct {
	Prompt      string
	Placeholder string
	OnComplete  func(string) tea.Cmd
}

type ConfirmationData struct {
	Message   string
	OnConfirm func(bool) tea.Cmd
}

// Helper functions for creating view switch commands
func SwitchToView(view ViewType, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return SwitchViewMsg{View: view, Data: data}
	}
}

func BackToPreviousView() tea.Cmd {
	return func() tea.Msg {
		return BackToPreviousViewMsg{}
	}
}

// Shared styles
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)
		
	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("green")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1)
		
	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("green"))
		
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)
		
	selectedIssueStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))
		
	unselectedIssueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))
		
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true)

	selectedActionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))   // White for selected action text

	unselectedActionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))    // Gray for unselected action text

	historyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))    // Gray for command history
)