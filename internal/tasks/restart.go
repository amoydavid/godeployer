package tasks

import (
	"fmt"

	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// RestartAppTask 表示重启应用的任务
type RestartAppTask struct{}

// NewRestartAppTask 创建新的重启应用任务
func NewRestartAppTask() task.Task {
	return &RestartAppTask{}
}

// Name 返回任务名称
func (t *RestartAppTask) Name() string {
	return "restart_app"
}

// Description 返回任务描述
func (t *RestartAppTask) Description() string {
	return "重启远程服务器上的应用"
}

// Execute 执行重启应用任务
func (t *RestartAppTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 从配置中获取重启命令
	taskCfg, ok := ctx.Config.Tasks["restart_app"]
	if ok && taskCfg.Remote != "" {
		return exec.RunRemoteCommand(taskCfg.Remote)
	}

	// 如果未配置重启命令，提示用户
	ctx.Logger.Warn("未在配置文件中找到重启命令。请在 YAML 配置文件中添加 restart_app 任务配置。")
	ctx.Logger.Info("示例配置:")
	ctx.Logger.Info("tasks:")
	ctx.Logger.Info("  restart_app:")
	ctx.Logger.Info("    remote: 'cd {{remote_dir}}/current && pm2 restart your-app-name'")

	return fmt.Errorf("重启失败：未配置重启命令")
}
