package registry

import (
	"fmt"
	"sync"

	"deployer/internal/task"
)

// TaskRegistry 管理已注册的任务
type TaskRegistry struct {
	tasks map[string]task.Task
	mu    sync.RWMutex
}

// NewTaskRegistry 创建新的任务注册表
func NewTaskRegistry() *TaskRegistry {
	return &TaskRegistry{
		tasks: make(map[string]task.Task),
	}
}

// Register 向注册表添加任务
func (r *TaskRegistry) Register(t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := t.Name()
	if _, exists := r.tasks[name]; exists {
		return fmt.Errorf("任务 '%s' 已经注册", name)
	}

	r.tasks[name] = t
	return nil
}

// Get 通过名称检索任务
func (r *TaskRegistry) Get(name string) (task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tasks[name]
	if !exists {
		return nil, fmt.Errorf("任务 '%s' 未找到", name)
	}

	return t, nil
}

// RegisterCustomTask 创建并注册自定义任务
func (r *TaskRegistry) RegisterCustomTask(name, description string, fn task.TaskFunc) error {
	return r.Register(task.NewCustomTask(name, description, fn))
}

// ListTasks 返回所有已注册任务的名称
func (r *TaskRegistry) ListTasks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tasks))
	for name := range r.tasks {
		names = append(names, name)
	}

	return names
}
