// File: internal/installer/parallel.go
// Purpose: Parallel task execution engine with concurrency limits and timeout control
// Problem: Sequential installation takes too long; need to run multiple tasks concurrently safely
// Role: Core execution engine that manages goroutines, semaphores, and task orchestration
// Usage: Create ParallelExecutor, call Execute() with list of tasks
// Design choices: Uses semaphore pattern for concurrency limits; context for timeouts; WaitGroup for synchronization
// Assumptions: Tasks are independent within parallel groups; file system operations are thread-safe

package installer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/rkinnovate/dev-setup/internal/config"
)

// ParallelExecutor runs tasks concurrently with limits and timeout control
// What: Manages concurrent task execution with configurable parallelism and timeouts
// Why: Installing tools sequentially wastes time; parallel execution cuts installation from 40min to 5min
type ParallelExecutor struct {
	maxConcurrent int
	timeout       time.Duration
	ui            UI
}

// UI interface defines methods for user feedback
// What: Contract for progress reporting and user notifications
// Why: Decouples execution logic from UI presentation; allows testing with mock UI
type UI interface {
	StartTask(name string)
	CompleteTask(name string)
	FailTask(name string, err error)
	Info(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// TaskResult contains the outcome of a task execution
// What: Captures task execution result including output and errors
// Why: Allows collecting results from concurrent tasks for reporting
type TaskResult struct {
	Task   config.Task
	Error  error
	Output string
	Duration time.Duration
}

// NewParallelExecutor creates a new ParallelExecutor
// What: Constructor for ParallelExecutor with configurable concurrency and timeout
// Why: Centralizes executor creation with sensible defaults
// Params: maxConcurrent - max simultaneous tasks (8 recommended), timeout - max time for all tasks, ui - UI for feedback
// Returns: Configured ParallelExecutor instance
// Example: executor := NewParallelExecutor(8, 5*time.Minute, ui)
func NewParallelExecutor(maxConcurrent int, timeout time.Duration, ui UI) *ParallelExecutor {
	return &ParallelExecutor{
		maxConcurrent: maxConcurrent,
		timeout:       timeout,
		ui:            ui,
	}
}

// Execute runs all tasks according to their parallel groups
// What: Main execution method that orchestrates task running with parallelism and ordering
// Why: Implements the core parallel execution logic that speeds up installation
// Params: tasks - slice of tasks to execute
// Returns: Error if any required task failed, nil if all succeeded
// Example: err := executor.Execute(stageTasks)
// Edge cases: Handles mixed sequential/parallel tasks; collects errors from failed tasks
func (p *ParallelExecutor) Execute(tasks []config.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	// Group tasks by parallel_group
	groups := p.groupTasks(tasks)

	var allResults []TaskResult
	var mu sync.Mutex // Protects allResults

	// Process each group
	for groupName, groupTasks := range groups {
		if groupName == "" {
			// Sequential tasks - run one at a time
			for _, task := range groupTasks {
				p.ui.StartTask(task.Name)
				result := p.executeTask(ctx, task)
				allResults = append(allResults, result)

				if result.Error != nil {
					p.ui.FailTask(task.Name, result.Error)
					if task.Required {
						return fmt.Errorf("required task failed: %s: %w", task.Name, result.Error)
					}
				} else if result.Output == "Skipped (condition not met)" {
					// Task was skipped due to condition, don't mark as complete
					p.ui.Info("  Skipped: %s", task.Name)
				} else {
					p.ui.CompleteTask(task.Name)
				}
			}
		} else {
			// Parallel tasks - run concurrently within group
			results := p.executeParallelGroup(ctx, groupTasks)

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()

			// Check for required task failures
			for _, result := range results {
				if result.Error != nil && result.Task.Required {
					return fmt.Errorf("required task failed: %s: %w", result.Task.Name, result.Error)
				}
			}
		}
	}

	// Report any non-required failures as warnings
	for _, result := range allResults {
		if result.Error != nil && !result.Task.Required {
			p.ui.Warning("Optional task failed: %s: %v", result.Task.Name, result.Error)
		}
	}

	return nil
}

// executeParallelGroup executes a group of tasks concurrently
// What: Runs multiple tasks simultaneously with concurrency limit (semaphore pattern)
// Why: Maximizes CPU/network utilization without overwhelming system
// Params: ctx - context for timeout control, tasks - tasks to run in parallel
// Returns: Slice of TaskResult for all tasks in group
func (p *ParallelExecutor) executeParallelGroup(ctx context.Context, tasks []config.Task) []TaskResult {
	var wg sync.WaitGroup
	results := make([]TaskResult, len(tasks))
	semaphore := make(chan struct{}, p.maxConcurrent)

	for i, task := range tasks {
		wg.Add(1)

		go func(index int, t config.Task) {
			defer wg.Done()

			// Acquire semaphore slot
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release slot

			// Execute task
			p.ui.StartTask(t.Name)
			result := p.executeTask(ctx, t)
			results[index] = result

			if result.Error != nil {
				p.ui.FailTask(t.Name, result.Error)
			} else if result.Output == "Skipped (condition not met)" {
				// Task was skipped due to condition, don't mark as complete
				p.ui.Info("  Skipped: %s", t.Name)
			} else {
				p.ui.CompleteTask(t.Name)
			}
		}(i, task)
	}

	wg.Wait()
	return results
}

// executeTask executes a single task with retries and condition checking
// What: Runs one task command with retry logic and optional condition check
// Why: Individual task execution with fault tolerance (retries) and conditional execution
// Params: ctx - context for timeout control, task - task to execute
// Returns: TaskResult with execution outcome
// Edge cases: Skips task if condition check fails; retries on failure if RetryCount > 0
func (p *ParallelExecutor) executeTask(ctx context.Context, task config.Task) TaskResult {
	startTime := time.Now()

	// Check condition if specified
	if task.Condition != "" {
		if !p.checkCondition(ctx, task.Condition) {
			return TaskResult{
				Task:     task,
				Error:    nil, // Not an error, just skipped
				Output:   "Skipped (condition not met)",
				Duration: time.Since(startTime),
			}
		}
	}

	// Execute with retries
	retries := task.RetryCount
	if retries == 0 {
		retries = 1 // At least one attempt
	}

	var lastErr error
	var output string

	for attempt := 0; attempt < retries; attempt++ {
		if attempt > 0 {
			p.ui.Info("  Retry %d/%d: %s", attempt, retries-1, task.Name)
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
		}

		// Create command with task-specific timeout or use context timeout
		taskCtx := ctx
		if task.Timeout > 0 {
			var cancel context.CancelFunc
			taskCtx, cancel = context.WithTimeout(ctx, task.Timeout)
			defer cancel()
		}

		cmd := exec.CommandContext(taskCtx, "bash", "-c", task.Command)
		outputBytes, err := cmd.CombinedOutput()
		output = string(outputBytes)

		if err == nil {
			return TaskResult{
				Task:     task,
				Error:    nil,
				Output:   output,
				Duration: time.Since(startTime),
			}
		}

		lastErr = err
	}

	// All retries failed
	return TaskResult{
		Task:     task,
		Error:    fmt.Errorf("%w: %s", lastErr, strings.TrimSpace(output)),
		Output:   output,
		Duration: time.Since(startTime),
	}
}

// checkCondition checks if a task condition is met
// What: Executes condition command and returns true if exit code is 0
// Why: Allows conditional task execution (e.g., skip if already installed)
// Params: ctx - context for timeout control, condition - shell command to check
// Returns: true if condition command exits with 0, false otherwise
// Example: checkCondition(ctx, "command -v brew >/dev/null") returns true if brew exists
func (p *ParallelExecutor) checkCondition(ctx context.Context, condition string) bool {
	cmd := exec.CommandContext(ctx, "bash", "-c", condition)
	err := cmd.Run()
	return err == nil
}

// groupTasks groups tasks by their parallel_group field
// What: Organizes tasks into map keyed by parallel group name
// Why: Separates sequential tasks (empty group) from parallel groups for correct execution order
// Params: tasks - slice of tasks to group
// Returns: Map of group name to tasks in that group (empty string = sequential)
// Example: {"" -> [task1, task2], "homebrew" -> [task3, task4, task5]}
func (p *ParallelExecutor) groupTasks(tasks []config.Task) map[string][]config.Task {
	groups := make(map[string][]config.Task)

	for _, task := range tasks {
		groupName := task.ParallelGroup
		groups[groupName] = append(groups[groupName], task)
	}

	return groups
}

// GetTaskStatistics returns execution statistics for completed tasks
// What: Calculates total time, success rate, and other metrics from task results
// Why: Provides performance insights and helps identify bottlenecks
// Params: results - slice of TaskResult from completed tasks
// Returns: Statistics struct with aggregated metrics
func GetTaskStatistics(results []TaskResult) TaskStatistics {
	stats := TaskStatistics{
		TotalTasks: len(results),
	}

	var totalDuration time.Duration

	for _, result := range results {
		totalDuration += result.Duration

		if result.Error != nil {
			stats.FailedTasks++
		} else {
			stats.SuccessfulTasks++
		}

		if result.Duration > stats.LongestTask {
			stats.LongestTask = result.Duration
			stats.LongestTaskName = result.Task.Name
		}
	}

	if stats.TotalTasks > 0 {
		stats.AverageDuration = totalDuration / time.Duration(stats.TotalTasks)
	}

	return stats
}

// TaskStatistics contains metrics about task execution
// What: Aggregated statistics from a set of executed tasks
// Why: Helps measure and optimize installation performance
type TaskStatistics struct {
	TotalTasks       int
	SuccessfulTasks  int
	FailedTasks      int
	AverageDuration  time.Duration
	LongestTask      time.Duration
	LongestTaskName  string
}
