package task

import (
	deployctx "deployer/internal/context"
)

// Task 表示可部署的任务
type Task interface {
	// Name 返回任务名称
	Name() string

	// Description 返回人类可读的描述
	Description() string

	// Execute 执行任务
	Execute(ctx *deployctx.DeployContext) error
}

// TaskFunc 是可以转换为任务的函数
type TaskFunc func(ctx *deployctx.DeployContext) error

// CustomTask 将 TaskFunc 包装为 Task
type CustomTask struct {
	name        string
	description string
	fn          TaskFunc
}

// NewCustomTask 创建新的自定义任务
func NewCustomTask(name, description string, fn TaskFunc) Task {
	return &CustomTask{
		name:        name,
		description: description,
		fn:          fn,
	}
}

func (t *CustomTask) Name() string {
	return t.name
}

func (t *CustomTask) Description() string {
	return t.description
}

func (t *CustomTask) Execute(ctx *deployctx.DeployContext) error {
	return t.fn(ctx)
}
