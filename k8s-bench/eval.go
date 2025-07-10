// Copyright 2024 kubelet-wuhrai Contributors
//
// Licensed under the kubelet-wuhrai Custom License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://github.com/st-lzh/kubelet-wuhrai/blob/main/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This software is based on kubectl-ai by Google Cloud Platform:
// https://github.com/GoogleCloudPlatform/kubectl-ai

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/st-lzh/kubelet-wuhrai/k8s-bench/pkg/model"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func runEvaluation(ctx context.Context, config EvalConfig) error {
	if config.OutputDir == "" {
		return fmt.Errorf("must set OutputDir")
	}

	tasks, err := loadTasks(config)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Fallback to sequential execution if concurrency is not set
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}

	// Create a channel for tasks to be processed
	type taskJob struct {
		taskID string
		task   Task
	}
	taskCh := make(chan taskJob, len(tasks))

	// Create a channel for collecting results
	resultsCh := make(chan model.TaskResult, len(tasks)*len(config.LLMConfigs))

	// Create a separate channel for errors
	errorsCh := make(chan error, config.Concurrency)

	// Load all tasks into the tasks channel
	for taskID, task := range tasks {
		taskCh <- taskJob{taskID: taskID, task: task}
	}
	close(taskCh)

	// Create a wait group to track all workers
	var wg sync.WaitGroup

	fmt.Printf("Running tasks with concurrency: %d\n", config.Concurrency)

	// Start workers based on concurrency setting
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for job := range taskCh {
				fmt.Printf("Worker %d: Evaluating task: %s\n", workerID, job.taskID)

				for _, llmConfig := range config.LLMConfigs {
					taskOutputDir := ""
					if config.OutputDir != "" {
						taskOutputDir = filepath.Join(config.OutputDir, job.taskID, llmConfig.ID)
						if err := os.MkdirAll(taskOutputDir, 0755); err != nil {
							errorsCh <- fmt.Errorf("creating directory %q: %w", taskOutputDir, err)
							return
						}
					}

					var log io.Writer
					if taskOutputDir != "" {
						logPath := filepath.Join(taskOutputDir, "log.txt")
						logFile, err := os.Create(logPath)
						if err != nil {
							errorsCh <- fmt.Errorf("creating log file %q: %w", logPath, err)
							return
						}
						defer logFile.Close()
						log = logFile
					}

					result := evaluateTask(ctx, config, job.taskID, job.task, llmConfig, log)

					if taskOutputDir != "" {
						if err := writeToYAMLFile(filepath.Join(taskOutputDir, "results.yaml"), result); err != nil {
							errorsCh <- fmt.Errorf("writing results to file: %w", err)
							return
						}
					}
					resultsCh <- result
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(resultsCh)
	close(errorsCh)

	// Check if there were any errors
	for err := range errorsCh {
		if err != nil {
			return err
		}
	}

	// Collect and print results
	var allResults []model.TaskResult
	for result := range resultsCh {
		allResults = append(allResults, result)
	}

	printResults(allResults)
	return nil
}

// writeToYAMLFile will encode the specified object as yaml, and write it to the file.
func writeToYAMLFile(p string, obj any) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshaling to yaml: %w", err)
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("writing to file %q: %w", p, err)
	}
	return nil
}

func loadTasks(config EvalConfig) (map[string]Task, error) {
	tasks := make(map[string]Task)

	entries, err := os.ReadDir(config.TasksDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		if config.TaskPattern != "" && !strings.Contains(taskID, config.TaskPattern) {
			continue
		}

		taskFile := filepath.Join(config.TasksDir, taskID, "task.yaml")

		data, err := os.ReadFile(taskFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read task file %s: %w", taskFile, err)
		}

		var task Task
		if err := yaml.Unmarshal(data, &task); err != nil {
			return nil, fmt.Errorf("failed to parse task file %s: %w", taskFile, err)
		}

		// Skip disabled tasks
		if task.Disabled {
			fmt.Printf("Skipping disabled task: %s\n", taskID)
			continue
		}

		tasks[taskID] = task
	}

	return tasks, nil
}

func evaluateTask(ctx context.Context, config EvalConfig, taskID string, task Task, llmConfig model.LLMConfig, log io.Writer) model.TaskResult {
	result := model.TaskResult{
		Task:      taskID,
		LLMConfig: llmConfig,
	}

	taskOutputDir := filepath.Join(config.OutputDir, taskID, llmConfig.ID)

	x := &TaskExecution{
		AgentBin:      config.AgentBin,
		kubeConfig:    config.KubeConfig,
		result:        &result,
		llmConfig:     llmConfig,
		log:           log,
		task:          &task,
		taskID:        taskID,
		taskOutputDir: taskOutputDir,
	}

	taskDir := filepath.Join(config.TasksDir, taskID)
	taskDirAbs, err := filepath.Abs(taskDir)
	if err != nil {
		result.Result = "fail"
		result.Error = err.Error()
		return result
	}
	taskDir = taskDirAbs
	x.taskDir = taskDir

	defer func() {
		if err := x.runCleanup(ctx); err != nil {
			fmt.Printf("Warning: cleanup failed for task %s: %v\n", taskID, err)
		}
	}()

	if err := x.runSetup(ctx); err != nil {
		// Unexpected error
		result.Error = err.Error()
		return result
	}

	// Run the agent
	if err := x.runAgent(ctx); err != nil {
		// Unexpected error
		result.Error = err.Error()
		return result
	}

	// Run verifier if specified
	if task.Verifier != "" {
		verifierPath := filepath.Join(taskDir, task.Verifier)
		cmd := exec.CommandContext(ctx, verifierPath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", x.kubeConfig))
		fmt.Printf("\nRunning verifier for task %s\n", taskID)

		err := x.runCommand(cmd)
		if err == nil {
			result.Result = "success"
		} else if _, ok := err.(*exec.ExitError); ok {
			// "Normal" script failure
			result.Result = "fail"
		} else {
			// Unexpected error
			result.Error = err.Error()
		}
	}

	return result
}

type TaskExecution struct {
	// kubeConfig is the path to the kubeconfig file we should use.
	// It will be created in IsolationModeCluster
	kubeConfig string

	// AgentBin holds the path to the agent to execute
	AgentBin string

	llmConfig model.LLMConfig
	result    *model.TaskResult
	log       io.Writer
	task      *Task
	taskID    string
	taskDir   string

	// taskOutputDir is where we can create artifacts or write logs while executing the task
	taskOutputDir string

	// cleanupFunctions are a set of cleanupFunctions we run to undo anything we ran
	cleanupFunctions []func() error
}

func (x *TaskExecution) runSetup(ctx context.Context) error {
	log := klog.FromContext(ctx)

	// Create cluster if requested
	if x.task.Isolation == IsolationModeCluster {
		kubeconfigPath := filepath.Join(x.taskDir, "kubeconfig.yaml")
		x.kubeConfig = kubeconfigPath

		clusterName := fmt.Sprintf("k8s-bench-%s", x.taskID)
		log.Info("creating kind cluster", "name", clusterName)

		args := []string{
			"kind",
			"create", "cluster",
			"--name", clusterName,
			"--wait", "5m",
			"--kubeconfig", kubeconfigPath,
		}
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = x.taskDir

		x.cleanupFunctions = append(x.cleanupFunctions, func() error {
			args := []string{
				"kind",
				"delete", "cluster",
				"--name", clusterName,
				"--kubeconfig", kubeconfigPath,
			}
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Dir = x.taskDir
			return x.runCommand(cmd)
		})

		if err := x.runCommand(cmd); err != nil {
			return err
		}
	}

	// Run setup if specified
	if x.task.Setup != "" {
		setupPath := filepath.Join(x.taskDir, x.task.Setup)
		cmd := exec.CommandContext(ctx, setupPath)
		cmd.Dir = x.taskDir
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", x.kubeConfig))

		if err := x.runCommand(cmd); err != nil {
			return err
		}
	}

	return nil
}

func (x *TaskExecution) runCleanup(ctx context.Context) error {
	var errs []error

	// Run cleanup if specified
	if x.task.Cleanup != "" {
		cleanupPath := filepath.Join(x.taskDir, x.task.Cleanup)
		cmd := exec.CommandContext(ctx, cleanupPath)
		cmd.Dir = x.taskDir
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", x.kubeConfig))

		if err := x.runCommand(cmd); err != nil {
			fmt.Printf("Warning: cleanup failed for task %s: %v\n", x.taskID, err)
		}
	}

	for _, cleanup := range x.cleanupFunctions {
		if err := cleanup(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (x *TaskExecution) runAgent(ctx context.Context) error {
	tracePath := filepath.Join(x.taskOutputDir, "trace.yaml")

	args := []string{
		"--kubeconfig", x.kubeConfig,
		"--llm-provider", x.llmConfig.ProviderID,
		fmt.Sprintf("--enable-tool-use-shim=%t", x.llmConfig.EnableToolUseShim),
		fmt.Sprintf("--quiet=%t", x.llmConfig.Quiet),
		"--model", x.llmConfig.ModelID,
		"--trace-path", tracePath,
		"--skip-permissions",
	}

	stdinReader, stdinWriter := io.Pipe()

	cmd := exec.CommandContext(ctx,
		x.AgentBin,
		args...,
	)
	cmd.Stdin = stdinReader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var stdoutBuffer bytes.Buffer
	if x.log != nil {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, x.log, &stdoutBuffer)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, x.log)
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", x.kubeConfig))

	go func() {
		// TODO: Wait for idle between sending steps?
		for _, step := range x.task.Script {
			fmt.Fprintf(stdinWriter, "%s\n", step.Prompt)
		}
		stdinWriter.Close()
	}()

	if err := cmd.Run(); err != nil {
		return err
	}

	// check any expectations
	for _, expect := range x.task.Expect {
		if expect.Contains != "" {
			// find the output after the last run command and search it
			agentOutput := stdoutBuffer.String()
			lastToolRunIndex := strings.LastIndex(agentOutput, "Running:")
			lastOutputIndex := strings.Index(agentOutput[lastToolRunIndex:], "\n")

			lastCmdOutput := agentOutput[lastToolRunIndex+lastOutputIndex+1:]
			if lastToolRunIndex == -1 {
				// if no tool run found, parse the entire output
				lastCmdOutput = agentOutput
			}
			if !strings.Contains(lastCmdOutput, expect.Contains) {
				x.result.AddFailure("expected value %q not found in output %q", expect.Contains, lastCmdOutput)
				return fmt.Errorf("expected value %q not found in agent output", expect.Contains)
			}
		}
	}

	return nil
}

func (x *TaskExecution) runCommand(cmd *exec.Cmd) error {
	fmt.Printf("\nRunning command: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if x.log != nil {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, x.log)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, x.log)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running command %v: %w", strings.Join(cmd.Args, " "), err)
	}
	return nil
}

func printResults(allResults []model.TaskResult) {
	fmt.Println("\nEvaluation Results:")
	fmt.Println("==================")

	for _, result := range allResults {
		fmt.Printf("\nTask: %s\n", result.Task)
		fmt.Printf("  LLM Config: %+v\n", result.LLMConfig)
		fmt.Printf("    %v\n", result.Result)
		if result.Error != "" {
			fmt.Printf("    Error: %s\n", result.Error)
		}
	}
}
