package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CLIExecutor handles execution of the actual taskmaster CLI commands
type CLIExecutor struct {
	cliPath string
}

// NewCLIExecutor creates a new CLI executor with the path to the taskmaster CLI
func NewCLIExecutor() *CLIExecutor {
	// Find the CLI script relative to the TUI binary
	cliPath := filepath.Join("..", "scripts", "dev.js")
	return &CLIExecutor{cliPath: cliPath}
}

// CLIResult represents the result of a CLI command execution
type CLIResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
	Error   string `json:"error"`
}

// ParsePRD executes the parse-prd command
func (e *CLIExecutor) ParsePRD(filePath, outputPath string, numTasks int, force, appendMode bool) CLIResult {
	args := []string{e.cliPath, "parse-prd", filePath, outputPath, fmt.Sprintf("--num-tasks=%d", numTasks)}
	
	if force {
		args = append(args, "--force")
	}
	if appendMode {
		args = append(args, "--append")
	}

	return e.executeCommand("node", args...)
}

// AddTask executes the add-task command
func (e *CLIExecutor) AddTask(filePath, prompt, title, description, details, testStrategy, dependencies, priority, taskType string, useResearch bool) CLIResult {
	args := []string{e.cliPath, "add-task", filePath}
	
	if prompt != "" {
		args = append(args, "--prompt", prompt)
	} else {
		args = append(args, "--title", title, "--description", description)
		if details != "" {
			args = append(args, "--details", details)
		}
		if testStrategy != "" {
			args = append(args, "--test-strategy", testStrategy)
		}
	}
	
	if dependencies != "" {
		args = append(args, "--dependencies", dependencies)
	}
	if priority != "" {
		args = append(args, "--priority", priority)
	}
	if taskType != "" {
		args = append(args, "--type", taskType)
	}
	if useResearch {
		args = append(args, "--research")
	}

	return e.executeCommand("node", args...)
}

// NextTask executes the next-task command
func (e *CLIExecutor) NextTask(filePath string) CLIResult {
	args := []string{e.cliPath, "next-task", filePath}
	return e.executeCommand("node", args...)
}

// ShowTask executes the show-task command
func (e *CLIExecutor) ShowTask(filePath, taskID string) CLIResult {
	args := []string{e.cliPath, "show-task", filePath, taskID}
	return e.executeCommand("node", args...)
}

// AddDependency executes the add-dependency command
func (e *CLIExecutor) AddDependency(filePath, taskID, dependencyID string) CLIResult {
	args := []string{e.cliPath, "add-dependency", filePath, taskID, dependencyID}
	return e.executeCommand("node", args...)
}

// UpdateTasks executes the update-tasks command
func (e *CLIExecutor) UpdateTasks(filePath, prompt string, taskIDs []string, useResearch bool) CLIResult {
	args := []string{e.cliPath, "update-tasks", filePath, "--prompt", prompt}
	
	if len(taskIDs) > 0 {
		args = append(args, "--task-ids", strings.Join(taskIDs, ","))
	}
	if useResearch {
		args = append(args, "--research")
	}

	return e.executeCommand("node", args...)
}

// UpdateOneTask executes the update-task command for a single task
func (e *CLIExecutor) UpdateOneTask(filePath, taskID, prompt string, useResearch bool) CLIResult {
	args := []string{e.cliPath, "update-task", filePath, taskID, "--prompt", prompt}
	
	if useResearch {
		args = append(args, "--research")
	}

	return e.executeCommand("node", args...)
}

// UpdateSubtask executes the update-subtask command
func (e *CLIExecutor) UpdateSubtask(filePath, taskID, subtaskID, prompt string, useResearch bool) CLIResult {
	args := []string{e.cliPath, "update-subtask", filePath, taskID, subtaskID, "--prompt", prompt}
	
	if useResearch {
		args = append(args, "--research")
	}

	return e.executeCommand("node", args...)
}

// GenerateTaskFiles executes the generate-task-files command
func (e *CLIExecutor) GenerateTaskFiles(filePath, outputDir string, force bool) CLIResult {
	args := []string{e.cliPath, "generate-task-files", filePath, outputDir}
	
	if force {
		args = append(args, "--force")
	}

	return e.executeCommand("node", args...)
}

// SetTaskStatus executes the set-task-status command
func (e *CLIExecutor) SetTaskStatus(filePath, taskID, status string) CLIResult {
	args := []string{e.cliPath, "set-task-status", filePath, taskID, status}
	return e.executeCommand("node", args...)
}

// ListTasks executes the list-tasks command
func (e *CLIExecutor) ListTasks(filePath, status, priority string, showSubtasks bool) CLIResult {
	args := []string{e.cliPath, "list-tasks", filePath}
	
	if status != "" {
		args = append(args, "--status", status)
	}
	if priority != "" {
		args = append(args, "--priority", priority)
	}
	if showSubtasks {
		args = append(args, "--show-subtasks")
	}

	return e.executeCommand("node", args...)
}

// ExpandTask executes the expand-task command
func (e *CLIExecutor) ExpandTask(filePath, taskID, prompt string, numSubtasks int, useResearch bool) CLIResult {
	args := []string{e.cliPath, "expand-task", filePath, taskID, "--prompt", prompt}
	
	if numSubtasks > 0 {
		args = append(args, fmt.Sprintf("--num-subtasks=%d", numSubtasks))
	}
	if useResearch {
		args = append(args, "--research")
	}

	return e.executeCommand("node", args...)
}

// AnalyzeComplexity executes the analyze-complexity command
func (e *CLIExecutor) AnalyzeComplexity(filePath string, threshold int, outputPath string) CLIResult {
	args := []string{e.cliPath, "analyze-complexity", filePath}
	
	if threshold > 0 {
		args = append(args, fmt.Sprintf("--threshold=%d", threshold))
	}
	if outputPath != "" {
		args = append(args, "--output", outputPath)
	}

	return e.executeCommand("node", args...)
}

// ClearSubtasks executes the clear-subtasks command
func (e *CLIExecutor) ClearSubtasks(filePath, taskID string) CLIResult {
	args := []string{e.cliPath, "clear-subtasks", filePath, taskID}
	return e.executeCommand("node", args...)
}

// executeCommand runs a command and returns the result
func (e *CLIExecutor) executeCommand(command string, args ...string) CLIResult {
	cmd := exec.Command(command, args...)
	
	// Set the working directory to the parent of the TUI directory
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = filepath.Join(wd, "..")
	}
	
	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	
	result := CLIResult{
		Output: string(output),
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Command failed: %s", err.Error())
	} else {
		result.Success = true
		result.Message = "Command executed successfully"
	}
	
	return result
}

// Global CLI executor instance
var cliExecutor = NewCLIExecutor()