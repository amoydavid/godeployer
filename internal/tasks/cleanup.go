package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
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
	cfg, ok := ctx.GetTaskConfig("cleanup")
	if ok {
		if cfg.Remote != "" {
			if err := exec.RunRemoteCommand(cfg.Remote); err != nil {
				return err
			}
		}

		if cfg.Local != "" {
			if err := exec.RunLocalCommand(cfg.Local); err != nil {
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

	// 清理本地临时文件（向后兼容：清理旧版本遗留的固定临时目录）
	legacyTempDir := ".deploy_tmp"
	if _, err := os.Stat(legacyTempDir); err == nil {
		// 目录存在，删除它
		if err := os.RemoveAll(legacyTempDir); err != nil {
			ctx.Logger.Warnf("Failed to clean legacy temp directory: %v", err)
		} else {
			ctx.Logger.Infof("Cleaned legacy temp directory: %s", legacyTempDir)
		}
	}

	// 清理系统临时目录中的 deployer 临时文件
	// 读取系统临时目录，查找并清理 deployer-* 目录
	systemTempDir := os.TempDir()
	entries, err := os.ReadDir(systemTempDir)
	if err != nil {
		ctx.Logger.Warnf("Failed to read system temp directory: %v", err)
		return nil
	}

	cleanedCount := 0
	for _, entry := range entries {
		// 检查是否是 deployer 创建的临时目录
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "deployer-") {
			tempPath := filepath.Join(systemTempDir, entry.Name())
			// 获取目录信息，检查是否过旧（可选，这里直接删除）
			if err := os.RemoveAll(tempPath); err != nil {
				ctx.Logger.Warnf("Failed to clean temp directory %s: %v", tempPath, err)
			} else {
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		ctx.Logger.Infof("Cleaned %d system temp directories", cleanedCount)
	}

	return nil
}
