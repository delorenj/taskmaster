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
	clearSubtasksFormKeyFile = "file"
	clearSubtasksFormKeyIDs  = "ids" // Comma-separated task IDs
	clearSubtasksFormKeyAll  = "all"
)

// ClearSubtasksModel holds the state for the clear subtasks form.
type ClearSubtasksModel struct {
	form         *huh.Form
	aborted      bool
	isProcessing bool
	statusMsg    string
	width        int

	// Form values
	FilePath string
	TaskIDs  string // Can be empty if 'AllTasks' is true
	AllTasks bool   // Clear subtasks from all tasks
}

// NewClearSubtasksForm creates a new form for the clear-subtasks command.
func NewClearSubtasksForm() *ClearSubtasksModel {
	m := &ClearSubtasksModel{
		AllTasks: false, // Default to not clearing all tasks
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key(clearSubtasksFormKeyFile).
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
				Key(clearSubtasksFormKeyIDs).
				Title("Task ID(s) (Optional)").
				Description("IDs of tasks to clear subtasks from. Leave empty if 'Clear All' is Yes.").
				Prompt("🆔 ").
				// Validation will be handled in the Update method based on 'AllTasks'
				Value(&m.TaskIDs),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Key(clearSubtasksFormKeyAll).
				Title("Clear Subtasks from All Tasks").
				Description("Clear subtasks from all tasks in the file?").
				Affirmative("Yes").
				Negative("No").
				Value(&m.AllTasks),
		),
	).WithTheme(huh.ThemeDracula())

	return m
}

func (m *ClearSubtasksModel) Init() tea.Cmd {
	m.isProcessing = false
	m.statusMsg = ""
	m.aborted = false
	return m.form.Init()
}

func (m *ClearSubtasksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.isProcessing {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" { return m, tea.Quit }
		}
		return m, nil
	}

	var cmds []tea.Cmd
	formModel, cmd := m.form.Update(msg)
	if updatedForm, ok := formModel.(*huh.Form); ok {
		m.form = updatedForm
	} else {
		m.statusMsg = "Error: Form update returned unexpected type."
		fmt.Fprintf(os.Stderr, "Critical Error: clear_subtasks_form.go - form update did not return *huh.Form. Got: %T\n", formModel)
		return m, tea.Quit
	}
	cmds = append(cmds, cmd)

	// Custom validation logic based on 'AllTasks'
	allTasksSelected := m.form.GetBool(clearSubtasksFormKeyAll)
	taskIDsProvided := m.form.GetString(clearSubtasksFormKeyIDs) != ""

	// Note: Direct field access for validation is not available in huh v0.7.0
	// We'll handle validation through the form's overall validation state
	if !allTasksSelected && !taskIDsProvided {
		if m.form.State == huh.StateCompleted { m.form.State = huh.StateNormal } // Prevent completion
	}


	if m.form.State == huh.StateCompleted {
		// Re-check validation before final processing because direct struct binding might occur before this point
		if !m.AllTasks && m.TaskIDs == "" { // m.AllTasks and m.TaskIDs are bound from form
			m.statusMsg = "Error: Task ID(s) are required if 'Clear All' is No."
			m.form.State = huh.StateNormal // Revert to allow correction
			return m, nil
		}

		m.statusMsg = "Executing clear-subtasks command..."
		m.isProcessing = true
		return m, m.executeClearSubtasksCommand()
	}

	if m.form.State == huh.StateAborted {
		m.aborted = true
		return m, func() tea.Msg { return backToMenuMsg{} }
	}

	switch msg := msg.(type) {
	case clearSubtasksCompleteMsg:
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
				m.aborted = true; return m, func() tea.Msg { return backToMenuMsg{} }
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	return m, tea.Batch(cmds...)
}

func (m *ClearSubtasksModel) View() string {
	if m.aborted { return "Form aborted. Returning to main menu..." }

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
	return lipgloss.NewStyle().Width(m.width).Padding(1, 2).Render(viewBuilder.String())
}

// GetFormValues retrieves the structured data after completion.
func (m *ClearSubtasksModel) GetFormValues() (map[string]interface{}, error) {
	if m.form.State != huh.StateCompleted {
		return nil, fmt.Errorf("form is not yet completed")
	}
	// Ensure TaskIDs is empty if AllTasks is true for command logic if needed
	taskIDsForCmd := m.TaskIDs
	if m.AllTasks {
		taskIDsForCmd = ""
	}
	return map[string]interface{}{
		clearSubtasksFormKeyFile: m.FilePath,
		clearSubtasksFormKeyIDs:  taskIDsForCmd,
		clearSubtasksFormKeyAll:  m.AllTasks,
	}, nil
}

// clearSubtasksCompleteMsg is sent when the command execution is complete
type clearSubtasksCompleteMsg struct {
	result CLIResult
}

// executeClearSubtasksCommand executes the actual clear-subtasks CLI command
// The CLI method expects a single taskID, so we'll handle multiple IDs by calling it for each one
func (m *ClearSubtasksModel) executeClearSubtasksCommand() tea.Cmd {
	return func() tea.Msg {
		if m.AllTasks {
			// For "all tasks", we would need a different approach
			// Since CLI expects a specific taskID, we'll return an error for now
			return clearSubtasksCompleteMsg{result: CLIResult{
				Success: false,
				Error:   "Clearing subtasks for all tasks is not yet supported via CLI",
			}}
		}
		
		// Parse task IDs from comma-separated string
		taskIDs := strings.Split(m.TaskIDs, ",")
		var results []string
		var hasError bool
		var lastError string
		
		for _, taskID := range taskIDs {
			trimmedID := strings.TrimSpace(taskID)
			if trimmedID == "" {
				continue
			}
			
			result := cliExecutor.ClearSubtasks(m.FilePath, trimmedID)
			if result.Success {
				results = append(results, fmt.Sprintf("✅ Task %s: %s", trimmedID, result.Output))
			} else {
				hasError = true
				lastError = result.Error
				results = append(results, fmt.Sprintf("❌ Task %s: %s", trimmedID, result.Error))
			}
		}
		
		if len(results) == 0 {
			return clearSubtasksCompleteMsg{result: CLIResult{
				Success: false,
				Error:   "No valid task IDs provided",
			}}
		}
		
		return clearSubtasksCompleteMsg{result: CLIResult{
			Success: !hasError,
			Error:   lastError,
			Output:  strings.Join(results, "\n"),
		}}
	}
}

var _ tea.Model = &ClearSubtasksModel{}
