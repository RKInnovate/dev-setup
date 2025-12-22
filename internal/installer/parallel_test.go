// File: internal/installer/parallel_test.go
// Purpose: Unit tests for parallel executor engine
// Problem: Need to verify parallel execution logic works correctly
// Role: Test suite for ParallelExecutor functionality
// Usage: Run with `go test ./internal/installer`
// Design choices: Uses table-driven tests; mocks UI interface; tests edge cases
// Assumptions: Test environment has bash available

package installer

import (
	"context"
	"testing"
	"time"

	"github.com/rkinnovate/dev-setup/internal/config"
)

// mockUI implements UI interface for testing
type mockUI struct {
	tasks []string
	calls []string
}

func (m *mockUI) StartTask(name string) {
	m.tasks = append(m.tasks, name)
	m.calls = append(m.calls, "StartTask:"+name)
}

func (m *mockUI) CompleteTask(name string) {
	m.calls = append(m.calls, "CompleteTask:"+name)
}

func (m *mockUI) FailTask(name string, err error) {
	m.calls = append(m.calls, "FailTask:"+name)
}

func (m *mockUI) Info(format string, args ...interface{}) {
	m.calls = append(m.calls, "Info")
}

func (m *mockUI) Warning(format string, args ...interface{}) {
	m.calls = append(m.calls, "Warning")
}

func (m *mockUI) Error(format string, args ...interface{}) {
	m.calls = append(m.calls, "Error")
}

func TestParallelExecutor_ExecuteSequential(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	tasks := []config.Task{
		{
			Name:    "Task 1",
			Command: "echo 'task1'",
		},
		{
			Name:    "Task 2",
			Command: "echo 'task2'",
		},
	}

	err := executor.Execute(tasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(ui.tasks) != 2 {
		t.Errorf("Expected 2 tasks started, got %d", len(ui.tasks))
	}
}

func TestParallelExecutor_ExecuteParallel(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	tasks := []config.Task{
		{
			Name:          "Task 1",
			Command:       "sleep 0.1 && echo 'task1'",
			ParallelGroup: "group1",
		},
		{
			Name:          "Task 2",
			Command:       "sleep 0.1 && echo 'task2'",
			ParallelGroup: "group1",
		},
	}

	start := time.Now()
	err := executor.Execute(tasks)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should complete in ~0.1s (parallel) not ~0.2s (sequential)
	if duration > 500*time.Millisecond {
		t.Errorf("Tasks took too long (%v), not running in parallel?", duration)
	}
}

func TestParallelExecutor_RequiredTaskFailure(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	tasks := []config.Task{
		{
			Name:     "Failing Task",
			Command:  "exit 1",
			Required: true,
		},
	}

	err := executor.Execute(tasks)
	if err == nil {
		t.Fatal("Expected error for required task failure, got nil")
	}
}

func TestParallelExecutor_OptionalTaskFailure(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	tasks := []config.Task{
		{
			Name:     "Optional Failing Task",
			Command:  "exit 1",
			Required: false,
		},
		{
			Name:    "Success Task",
			Command: "echo 'success'",
		},
	}

	err := executor.Execute(tasks)
	if err != nil {
		t.Fatalf("Expected no error for optional task failure, got: %v", err)
	}
}

func TestParallelExecutor_Condition(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	tasks := []config.Task{
		{
			Name:      "Conditional Task",
			Command:   "echo 'should not run'",
			Condition: "false",  // Condition fails
		},
	}

	err := executor.Execute(tasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Task should be skipped, so no completion call
	hasComplete := false
	for _, call := range ui.calls {
		if call == "CompleteTask:Conditional Task" {
			hasComplete = true
		}
	}
	if hasComplete {
		t.Error("Task should have been skipped due to condition")
	}
}

func TestParallelExecutor_Retry(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	// Create a temp file to track retry attempts
	tasks := []config.Task{
		{
			Name:       "Retry Task",
			Command:    "exit 1",  // Always fails
			Required:   false,
			RetryCount: 2,
		},
	}

	err := executor.Execute(tasks)
	if err != nil {
		t.Fatalf("Expected no error for optional task, got: %v", err)
	}

	// Should see Info calls for retries
	infoCount := 0
	for _, call := range ui.calls {
		if call == "Info" {
			infoCount++
		}
	}

	// Should have at least one retry info message
	if infoCount == 0 {
		t.Error("Expected retry info messages")
	}
}

func TestParallelExecutor_Timeout(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 1*time.Second, ui)

	tasks := []config.Task{
		{
			Name:     "Slow Task",
			Command:  "sleep 10",  // Takes too long
			Required: true,
		},
	}

	err := executor.Execute(tasks)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestGroupTasks(t *testing.T) {
	ui := &mockUI{}
	executor := NewParallelExecutor(4, 30*time.Second, ui)

	tasks := []config.Task{
		{Name: "Sequential 1", ParallelGroup: ""},
		{Name: "Parallel 1", ParallelGroup: "group1"},
		{Name: "Parallel 2", ParallelGroup: "group1"},
		{Name: "Sequential 2", ParallelGroup: ""},
		{Name: "Parallel 3", ParallelGroup: "group2"},
	}

	groups := executor.groupTasks(tasks)

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}

	if len(groups[""]) != 2 {
		t.Errorf("Expected 2 sequential tasks, got %d", len(groups[""]))
	}

	if len(groups["group1"]) != 2 {
		t.Errorf("Expected 2 tasks in group1, got %d", len(groups["group1"]))
	}

	if len(groups["group2"]) != 1 {
		t.Errorf("Expected 1 task in group2, got %d", len(groups["group2"]))
	}
}

func TestGetTaskStatistics(t *testing.T) {
	results := []TaskResult{
		{
			Task:     config.Task{Name: "Task 1"},
			Error:    nil,
			Duration: 100 * time.Millisecond,
		},
		{
			Task:     config.Task{Name: "Task 2"},
			Error:    nil,
			Duration: 200 * time.Millisecond,
		},
		{
			Task:     config.Task{Name: "Task 3"},
			Error:    context.DeadlineExceeded,
			Duration: 50 * time.Millisecond,
		},
	}

	stats := GetTaskStatistics(results)

	if stats.TotalTasks != 3 {
		t.Errorf("Expected 3 total tasks, got %d", stats.TotalTasks)
	}

	if stats.SuccessfulTasks != 2 {
		t.Errorf("Expected 2 successful tasks, got %d", stats.SuccessfulTasks)
	}

	if stats.FailedTasks != 1 {
		t.Errorf("Expected 1 failed task, got %d", stats.FailedTasks)
	}

	if stats.LongestTaskName != "Task 2" {
		t.Errorf("Expected longest task to be 'Task 2', got '%s'", stats.LongestTaskName)
	}

	expectedAvg := (100*time.Millisecond + 200*time.Millisecond + 50*time.Millisecond) / 3
	if stats.AverageDuration != expectedAvg {
		t.Errorf("Expected average duration %v, got %v", expectedAvg, stats.AverageDuration)
	}
}
