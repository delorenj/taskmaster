package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	showTaskFormKeyFile         = "file"
	showTaskFormKeyID           = "id"
	showTaskFormKeyStatusFilter = "status-filter" // For subtasks
)

// Using FilterStatus from list_tasks_form.go (assuming it's in the same package `main`)
// If not, it would need to be redefined or imported if in a different package.
// For this example, assuming FilterStatus (none, todo, in-progress, review, done) is available.

// ShowTaskModel holds the state for the show task form.
type ShowTaskModel struct {
	form         *huh.Form
	aborted      bool
	isProcessing bool // To simulate action, though 'show' might just display info
	statusMsg    string
	width        int

	// Form values
	FilePath      string
	TaskID        string
	StatusFilter  FilterStatus // For subtask filtering
}

// NewShowTaskForm creates a new form for the show task command.
func NewShowTaskForm() *ShowTaskModel {
	m := &ShowTaskModel{
		StatusFilter: FilterStatusNone, // Default to no filter for subtasks
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key(showTaskFormKeyFile).
				Title("Tasks File Path").
				Description("Path to the tasks file (e.g., tasks.md).").
				Prompt("📄 ").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("tasks file path cannot be empty")
					}
					return nil
				}).
				Value(&m.FilePath),

			huh.NewInput().
				Key(showTaskFormKeyID).
				Title("Task ID").
				Description("ID of the task to show (e.g., \"1\", \"2.1\").").
				Prompt("🆔 ").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("task ID cannot be empty")
					}
					return nil
				}).
				Value(&m.TaskID),
		),
		huh.NewGroup(
			huh.NewSelect[FilterStatus](). // Using FilterStatus from list_tasks_form.go
				Key(showTaskFormKeyStatusFilter).
				Title("Subtask Status Filter (Optional)").
				Description("Filter subtasks by status, or 'None' to show all.").
				Options(
					huh.NewOption("None (Show All Subtasks)", FilterStatusNone),
					huh.NewOption("To Do", FilterStatusTodo),
					huh.NewOption("In Progress", FilterStatusInProgress),
					huh.NewOption("Review", FilterStatusReview),
					huh.NewOption("Done", FilterStatusDone),
				).
				Value(&m.StatusFilter),
		),
	).WithTheme(huh.ThemeDracula())

	return m
}

func (m *ShowTaskModel) Init() tea.Cmd {
	m.isProcessing = false
	m.statusMsg = ""
	m.aborted = false
	return m.form.Init()
}

func (m *ShowTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.isProcessing {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
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
		m.statusMsg = "Error: Form update returned unexpected type."
		fmt.Fprintf(os.Stderr, "Critical Error: show_task_form.go - form update did not return *huh.Form. Got: %T\n", formModel)
		return m, tea.Quit
	}
	cmds = append(cmds, cmd)

	if m.form.State == huh.StateCompleted {
		m.statusMsg = "Executing show-task command..."
		m.isProcessing = true
		return m, m.executeShowTaskCommand()
	}

	if m.form.State == huh.StateAborted {
		m.aborted = true
		return m, func() tea.Msg { return backToMenuMsg{} }
	}

	switch msg := msg.(type) {
	case showTaskCompleteMsg:
		m.isProcessing = false
		if msg.result.Success {
			m.statusMsg = fmt.Sprintf("✅ Success!\n\n%s", msg.result.Output)
		} else {
			m.statusMsg = fmt.Sprintf("❌ Error: %s\n\n%s", msg.result.Error, msg.result.Output)
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

func (m *ShowTaskModel) View() string {
	if m.aborted {
		return "Form aborted. Returning to main menu..."
	}

	var viewBuilder strings.Builder
	viewBuilder.WriteString(m.form.View())

	if m.statusMsg != "" {
		viewBuilder.WriteString("\n\n")
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		if strings.HasPrefix(m.statusMsg, "Error:") {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}
		viewBuilder.WriteString(statusStyle.Render(m.statusMsg))
	}

	helpStyle := lipgloss.NewStyle().Faint(true)
	if m.isProcessing {
		viewBuilder.WriteString(helpStyle.Render("\n\nProcessing... Press Ctrl+C to force quit."))
	} else if m.form.State == huh.StateCompleted && strings.HasPrefix(m.statusMsg, "✅") {
		viewBuilder.WriteString(helpStyle.Render("\n\nCommand completed! Press Esc to return to main menu."))
	} else if m.form.State != huh.StateCompleted && m.form.State != huh.StateAborted {
		viewBuilder.WriteString(helpStyle.Render("\n\nPress Esc to return to main menu, Ctrl+C to quit application."))
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(1, 2).
		Render(viewBuilder.String())
}

// GetFormValues retrieves the structured data after completion.
func (m *ShowTaskModel) GetFormValues() (map[string]interface{}, error) {
	if m.form.State != huh.StateCompleted {
		return nil, fmt.Errorf("form is not yet completed")
	}
	return map[string]interface{}{
		showTaskFormKeyFile:         m.FilePath,
		showTaskFormKeyID:           m.TaskID,
		showTaskFormKeyStatusFilter: m.StatusFilter,
	}, nil
}

// showTaskCompleteMsg is sent when the command execution is complete
type showTaskCompleteMsg struct {
	result CLIResult
}

// executeShowTaskCommand executes the actual show-task CLI command
// Note: The CLI doesn't support status filtering for subtasks, so we ignore the StatusFilter field
func (m *ShowTaskModel) executeShowTaskCommand() tea.Cmd {
	return func() tea.Msg {
		result := cliExecutor.ShowTask(m.FilePath, m.TaskID)
		return showTaskCompleteMsg{result: result}
	}
}

var _ tea.Model = &ShowTaskModel{}
