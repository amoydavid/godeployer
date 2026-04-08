package tasks

import (
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// SymlinkReleaseTask 表示将当前发布版本软链到current的任务
type SymlinkReleaseTask struct{}

// NewSymlinkReleaseTask 创建新的软链任务
func NewSymlinkReleaseTask() task.Task {
	return &SymlinkReleaseTask{}
}

// Name 返回任务名称
func (t *SymlinkReleaseTask) Name() string {
	return "symlink_release"
}

// Description 返回任务描述
func (t *SymlinkReleaseTask) Description() string {
	return "将当前发布版本软链到current目录"
}

// Execute 执行软链任务
func (t *SymlinkReleaseTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 创建软链接到新的发布版本
	symlinkCmd := "ln -sfn {{release_path}} {{remote_dir}}/current"
	if err := exec.RunRemoteCommand(symlinkCmd); err != nil {
		return err
	}

	ctx.Logger.Success("Symlink created successfully, current now points to the latest release")
	return nil
}
