package localmodel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

// TaskManager manages tasks for local AI models
type TaskManager struct {
	tasks      map[string]*TaskRecord
	taskMutex  sync.RWMutex
	serverMgr  *ServerManager
	taskIDCounter int64
}

// NewTaskManager creates a new task manager
func NewTaskManager(serverMgr *ServerManager) *TaskManager {
	return &TaskManager{
		tasks:     make(map[string]*TaskRecord),
		serverMgr: serverMgr,
	}
}

// SubmitTask submits a new task
func (tm *TaskManager) SubmitTask(task *TaskRecord) (string, error) {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	// Generate task ID
	tm.taskIDCounter++
	task.ID = tm.taskIDCounter
	taskID := fmt.Sprintf("task-%d", task.ID)

	// Set default values
	if task.Status == "" {
		task.Status = TaskQueue
	}
	if task.StartTime == 0 {
		task.StartTime = time.Now().Unix()
	}

	// Store task
	tm.tasks[taskID] = task

	return taskID, nil
}

// GetTask returns a task by ID
func (tm *TaskManager) GetTask(taskID string) (*TaskRecord, error) {
	tm.taskMutex.RLock()
	defer tm.taskMutex.RUnlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	// Return a copy to prevent external modification
	taskCopy := *task
	return &taskCopy, nil
}

// UpdateTask updates a task
func (tm *TaskManager) UpdateTask(taskID string, updates map[string]interface{}) error {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "status":
			if status, ok := value.(string); ok {
				task.Status = TaskStatus(status)
			}
		case "statusMsg":
			if msg, ok := value.(string); ok {
				task.StatusMsg = msg
			}
		case "endTime":
			if endTime, ok := value.(int64); ok {
				task.EndTime = &endTime
			}
		case "result":
			if result, ok := value.(map[string]interface{}); ok {
				task.Result = result
			}
		case "jobResult":
			if jobResult, ok := value.(map[string]interface{}); ok {
				task.JobResult = jobResult
			}
		}
	}

	return nil
}

// ListTasks lists all tasks
func (tm *TaskManager) ListTasks() []TaskRecord {
	tm.taskMutex.RLock()
	defer tm.taskMutex.RUnlock()

	tasks := make([]TaskRecord, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		// Return copies to prevent external modification
		taskCopy := *task
		tasks = append(tasks, taskCopy)
	}

	return tasks
}

// DeleteTask deletes a task
func (tm *TaskManager) DeleteTask(taskID string) error {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	_, exists := tm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	delete(tm.tasks, taskID)
	return nil
}

// RunSoundTtsTask runs a sound TTS task
func (tm *TaskManager) RunSoundTtsTask(taskID string, serverKey string, text string, param map[string]interface{}) (*TaskResult, error) {
	// Update task status
	err := tm.UpdateTask(taskID, map[string]interface{}{
		"status": TaskRunning,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %v", err)
	}

	// Prepare function data
	funcData := ServerFunctionDataType{
		ID:    taskID,
		Result: make(map[string]interface{}),
		Param:  param,
		Text:   text,
	}

	// Call server function
	result, err := tm.serverMgr.CallFunction(serverKey, "soundTts", funcData)
	if err != nil {
		// Update task status to failed
		tm.UpdateTask(taskID, map[string]interface{}{
			"status":    TaskFail,
			"statusMsg": err.Error(),
		})
		return nil, fmt.Errorf("failed to call soundTts function: %v", err)
	}

	// Update task with result
	endTime := time.Now().Unix()
	err = tm.UpdateTask(taskID, map[string]interface{}{
		"status":  TaskSuccess,
		"endTime": endTime,
		"result":  result.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task with result: %v", err)
	}

	return result, nil
}

// RunSoundCloneTask runs a sound clone task
func (tm *TaskManager) RunSoundCloneTask(taskID string, serverKey string, text, promptAudio, promptText string, param map[string]interface{}) (*TaskResult, error) {
	// Update task status
	err := tm.UpdateTask(taskID, map[string]interface{}{
		"status": TaskRunning,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %v", err)
	}

	// Prepare function data
	funcData := ServerFunctionDataType{
		ID:          taskID,
		Result:      make(map[string]interface{}),
		Param:       param,
		Text:        text,
		PromptAudio: promptAudio,
		PromptText:  promptText,
	}

	// Call server function
	result, err := tm.serverMgr.CallFunction(serverKey, "soundClone", funcData)
	if err != nil {
		// Update task status to failed
		tm.UpdateTask(taskID, map[string]interface{}{
			"status":    TaskFail,
			"statusMsg": err.Error(),
		})
		return nil, fmt.Errorf("failed to call soundClone function: %v", err)
	}

	// Update task with result
	endTime := time.Now().Unix()
	err = tm.UpdateTask(taskID, map[string]interface{}{
		"status":  TaskSuccess,
		"endTime": endTime,
		"result":  result.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task with result: %v", err)
	}

	return result, nil
}

// RunVideoGenTask runs a video generation task
func (tm *TaskManager) RunVideoGenTask(taskID string, serverKey string, video, audio string, param map[string]interface{}) (*TaskResult, error) {
	// Update task status
	err := tm.UpdateTask(taskID, map[string]interface{}{
		"status": TaskRunning,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %v", err)
	}

	// Prepare function data
	funcData := ServerFunctionDataType{
		ID:    taskID,
		Result: make(map[string]interface{}),
		Param:  param,
		Video:  video,
		Audio:  audio,
	}

	// Call server function
	result, err := tm.serverMgr.CallFunction(serverKey, "videoGen", funcData)
	if err != nil {
		// Update task status to failed
		tm.UpdateTask(taskID, map[string]interface{}{
			"status":    TaskFail,
			"statusMsg": err.Error(),
		})
		return nil, fmt.Errorf("failed to call videoGen function: %v", err)
	}

	// Update task with result
	endTime := time.Now().Unix()
	err = tm.UpdateTask(taskID, map[string]interface{}{
		"status":  TaskSuccess,
		"endTime": endTime,
		"result":  result.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task with result: %v", err)
	}

	return result, nil
}

// RunAsrTask runs an ASR task
func (tm *TaskManager) RunAsrTask(taskID string, serverKey string, audio string, param map[string]interface{}) (*TaskResult, error) {
	// Update task status
	err := tm.UpdateTask(taskID, map[string]interface{}{
		"status": TaskRunning,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task status: %v", err)
	}

	// Prepare function data
	funcData := ServerFunctionDataType{
		ID:    taskID,
		Result: make(map[string]interface{}),
		Param:  param,
		Audio:  audio,
	}

	// Call server function
	result, err := tm.serverMgr.CallFunction(serverKey, "asr", funcData)
	if err != nil {
		// Update task status to failed
		tm.UpdateTask(taskID, map[string]interface{}{
			"status":    TaskFail,
			"statusMsg": err.Error(),
		})
		return nil, fmt.Errorf("failed to call asr function: %v", err)
	}

	// Update task with result
	endTime := time.Now().Unix()
	err = tm.UpdateTask(taskID, map[string]interface{}{
		"status":  TaskSuccess,
		"endTime": endTime,
		"result":  result.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task with result: %v", err)
	}

	return result, nil
}

// SaveTasks saves tasks to a file
func (tm *TaskManager) SaveTasks(filename string) error {
	tm.taskMutex.RLock()
	defer tm.taskMutex.RUnlock()

	data, err := json.Marshal(tm.tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write tasks to file: %v", err)
	}

	return nil
}

// LoadTasks loads tasks from a file
func (tm *TaskManager) LoadTasks(filename string) error {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read tasks from file: %v", err)
	}

	err = json.Unmarshal(data, &tm.tasks)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tasks: %v", err)
	}

	// Update task ID counter
	for _, task := range tm.tasks {
		if task.ID > tm.taskIDCounter {
			tm.taskIDCounter = task.ID
		}
	}

	return nil
}
