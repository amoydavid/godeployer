package tasks

import (
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// SetupTask 表示设置远程服务器部署目录的任务
type SetupTask struct{}

// NewSetupTask 创建新的设置任务
func NewSetupTask() task.Task {
	return &SetupTask{}
}

// Name 返回任务名称
func (t *SetupTask) Name() string {
	return "setup"
}

// Description 返回任务描述
func (t *SetupTask) Description() string {
	return "设置远程服务器上的部署目录结构"
}

// Execute 执行设置任务
func (t *SetupTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 如果配置中有自定义远程命令，则使用它
	taskCfg, ok := ctx.Config.Tasks["setup"]
	if ok && taskCfg.Remote != "" {
		return exec.RunRemoteCommand(taskCfg.Remote)
	}

	// 否则，使用默认设置命令
	// remoteDir := ctx.StageConfig.RemoteDir

	// 创建基本目录结构
	setupCmd := `
mkdir -p {{remote_dir}}/releases
mkdir -p {{remote_dir}}/shared
`

	return exec.RunRemoteCommand(setupCmd)
}
