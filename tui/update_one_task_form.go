package main

import (
	"fmt"
	"os"
	"strings"
	// "strconv" // Not strictly needed if ID is treated as string, but good for validation if numeric

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	updateOneTaskFormKeyFile     = "file"
	updateOneTaskFormKeyID       = "id"
	updateOneTaskFormKeyPrompt   = "prompt"
	updateOneTaskFormKeyResearch = "research"
)

// UpdateSingleTaskModel holds the state for the update-task (single task) form.
type UpdateSingleTaskModel struct {
	form         *huh.Form
	aborted      bool
	isProcessing bool
	status       string
	width        int

	// Form values
	FilePath string
	TaskID   string // Task ID can be string to accommodate various ID formats (e.g., alphanumeric)
	Prompt   string
	Research bool
}

// NewUpdateSingleTaskForm creates a new form for the update-task command.
func NewUpdateSingleTaskForm() *UpdateSingleTaskModel {
	m := &UpdateSingleTaskModel{
		Research: false, // Default for research
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key(updateOneTaskFormKeyFile).
				Title("Tasks File Path").
				Description("Path to the tasks file (e.g., tasks.md).").
				Prompt("📄 ").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("file path cannot be empty")
					}
					return nil
				}).
				Value(&m.FilePath),

			huh.NewInput().
				Key(updateOneTaskFormKeyID).
				Title("Task ID").
				Description("ID of the task to update.").
				Prompt("🆔 ").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("task ID cannot be empty")
					}
					// Add more specific ID validation if needed (e.g., numeric, specific format)
					return nil
				}).
				Value(&m.TaskID),

			huh.NewText().
				Key(updateOneTaskFormKeyPrompt).
				Title("Update Prompt").
				Description("Explain the changes for this specific task.").
				CharLimit(500). // Optional character limit
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("prompt cannot be empty")
					}
					return nil
				}).
				Value(&m.Prompt),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Key(updateOneTaskFormKeyResearch).
				Title("Use Research").
				Description("Incorporate research-backed updates for this task?").
				Affirmative("Yes").
				Negative("No").
				Value(&m.Research),
		),
	).WithTheme(huh.ThemeDracula())

	return m
}

func (m *UpdateSingleTaskModel) Init() tea.Cmd {
	m.isProcessing = false
	m.status = ""
	m.aborted = false
	return m.form.Init()
}

func (m *UpdateSingleTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.isProcessing {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
		return m, nil
	}

	var cmds []tea.Cmd
	formModel, cmd := m.form.Update(msg)
	if updatedForm, ok := formModel.(*huh.Form); ok {
		m.form = updatedForm
	} else {
		m.status = "Error: Form update returned unexpected type."
		fmt.Fprintf(os.Stderr, "Critical Error: update_one_task_form.go - form update did not return *huh.Form. Got: %T\n", formModel)
		return m, tea.Quit
	}
	cmds = append(cmds, cmd)

	if m.form.State == huh.StateCompleted {
		// Values are already bound to m.FilePath, m.TaskID, m.Prompt, m.Research.
		m.status = "Executing update-task command..."
		m.isProcessing = true
		return m, m.executeUpdateOneTaskCommand()
	}

	if m.form.State == huh.StateAborted {
		m.aborted = true
		return m, func() tea.Msg { return backToMenuMsg{} }
	}

	switch msg := msg.(type) {
	case updateOneTaskCompleteMsg:
		m.isProcessing = false
		if msg.result.Success {
			m.status = fmt.Sprintf("✅ Success!\n\n%s", msg.result.Output)
		} else {
			m.status = fmt.Sprintf("❌ Error: %s\n\n%s", msg.result.Error, msg.result.Output)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if !m.isProcessing {
				m.aborted = true
				return m, func() tea.Msg { return backToMenuMsg{} }
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	return m, tea.Batch(cmds...)
}

func (m *UpdateSingleTaskModel) View() string {
	if m.aborted {
		return "Form aborted. Returning to main menu..."
	}

	var viewBuilder strings.Builder
	viewBuilder.WriteString(m.form.View())

	if m.status != "" {
		viewBuilder.WriteString("\n\n")
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		if strings.HasPrefix(m.status, "Error:") {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}
		viewBuilder.WriteString(statusStyle.Render(m.status))
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	if m.isProcessing {
		viewBuilder.WriteString(helpStyle.Render("\n\nProcessing... Press Ctrl+C to force quit."))
	} else if m.form.State == huh.StateCompleted && strings.HasPrefix(m.status, "✅") {
		viewBuilder.WriteString(helpStyle.Render("\n\nCommand completed! Press Esc to return to main menu."))
	} else if m.form.State != huh.StateCompleted && m.form.State != huh.StateAborted {
		viewBuilder.WriteString(helpStyle.Render("\n\nPress Esc to return to main menu, Ctrl+C to quit application."))
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Render(viewBuilder.String())
}

// GetFormValues can be used to retrieve the structured data after completion.
func (m *UpdateSingleTaskModel) GetFormValues() (map[string]interface{}, error) {
	if m.form.State != huh.StateCompleted {
		return nil, fmt.Errorf("form is not yet completed")
	}
	return map[string]interface{}{
		updateOneTaskFormKeyFile:     m.FilePath,
		updateOneTaskFormKeyID:       m.TaskID,
		updateOneTaskFormKeyPrompt:   m.Prompt,
		updateOneTaskFormKeyResearch: m.Research,
	}, nil
}

// updateOneTaskCompleteMsg is sent when the command execution is complete
type updateOneTaskCompleteMsg struct {
	result CLIResult
}

// executeUpdateOneTaskCommand executes the actual update-task CLI command
func (m *UpdateSingleTaskModel) executeUpdateOneTaskCommand() tea.Cmd {
	return func() tea.Msg {
		result := cliExecutor.UpdateOneTask(m.FilePath, m.TaskID, m.Prompt, m.Research)
		return updateOneTaskCompleteMsg{result: result}
	}
}

// Ensure UpdateSingleTaskModel implements tea.Model.
var _ tea.Model = &UpdateSingleTaskModel{}
