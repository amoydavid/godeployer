package tasks

import (
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
	"fmt"
)

// CleanupTask 表示清理旧发布版本的任务
type CleanupTask struct{}

// NewCleanupTask 创建新的清理任务
func NewCleanupTask() task.Task {
	return &CleanupTask{}
}

// Name 返回任务名称
func (t *CleanupTask) Name() string {
	return "cleanup"
}

// Description 返回任务描述
func (t *CleanupTask) Description() string {
	return "清理旧的发布版本"
}

// Execute 执行清理任务
func (t *CleanupTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 如果配置中有自定义命令，使用它
	taskCfg, ok := ctx.Config.Tasks["cleanup"]
	if ok {
		if taskCfg.Remote != "" {
			if err := exec.RunRemoteCommand(taskCfg.Remote); err != nil {
				return err
			}
		}

		if taskCfg.Local != "" {
			if err := exec.RunLocalCommand(taskCfg.Local); err != nil {
				return err
			}
		}

		return nil
	}

	// 否则，使用默认清理命令
	keepReleases := ctx.StageConfig.KeepReleases
	if keepReleases <= 0 {
		keepReleases = 5 // 默认保留5个版本
	}

	// 删除旧的发布版本，保留最近的N个
	cleanupCmd := fmt.Sprintf("ls -dt %s/releases/* | tail -n +%d | xargs rm -rf || true",
		ctx.StageConfig.RemoteDir, keepReleases+1)

	// 在远程服务器上执行清理
	if err := exec.RunRemoteCommand(cleanupCmd); err != nil {
		return err
	}

	// 清理本地临时文件
	return exec.RunLocalCommand("rm -rf .deploy_tmp")
}
