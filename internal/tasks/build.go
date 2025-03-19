package tasks

import (
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// BuildTask 表示构建应用的任务
type BuildTask struct{}

// NewBuildTask 创建新的构建任务
func NewBuildTask() task.Task {
	return &BuildTask{}
}

// Name 返回任务名称
func (t *BuildTask) Name() string {
	return "build"
}

// Description 返回任务描述
func (t *BuildTask) Description() string {
	return "在本地构建应用"
}

// Execute 执行构建任务
func (t *BuildTask) Execute(ctx *deployctx.DeployContext) error {
	// 检查配置中是否有自定义构建命令
	taskCfg, ok := ctx.Config.Tasks["build"]
	if !ok || taskCfg.Local == "" {
		ctx.Logger.Info("没有配置构建命令，跳过构建")
		return nil
	}

	// 执行构建命令
	exec := executor.NewExecutor(ctx)
	return exec.RunLocalCommand(taskCfg.Local)
}
