package tasks

import (
	"fmt"

	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// InstallDependenciesTask 表示安装依赖的任务
type InstallDependenciesTask struct{}

// NewInstallDependenciesTask 创建新的安装依赖任务
func NewInstallDependenciesTask() task.Task {
	return &InstallDependenciesTask{}
}

// Name 返回任务名称
func (t *InstallDependenciesTask) Name() string {
	return "install_dependencies"
}

// Description 返回任务描述
func (t *InstallDependenciesTask) Description() string {
	return "在远程服务器安装应用依赖"
}

// Execute 执行安装依赖任务
func (t *InstallDependenciesTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 如果配置中有自定义命令，使用它
	cmd, ok := ctx.GetTaskRemoteCmd("install_dependencies")
	if ok {
		return exec.RunRemoteCommand(cmd)
	}

	// 否则，尝试检测项目类型并使用默认命令
	// 检查是否存在 package.json（Node.js 项目）
	checkNodeCmd := "if [ -f {{release_path}}/package.json ]; then echo 'node'; else echo 'unknown'; fi"
	output, err := exec.RunRemoteCommandWithOutput(checkNodeCmd)
	if err != nil {
		return fmt.Errorf("检查项目类型失败: %w", err)
	}

	if output == "node" {
		// 对于 Node.js 项目，使用 npm 或 yarn 或 pnpm
		checkYarnCmd := "if command -v yarn >/dev/null 2>&1; then echo 'yarn'; elif command -v pnpm >/dev/null 2>&1; then echo 'pnpm'; else echo 'npm'; fi"
		pkgManager, err := exec.RunRemoteCommandWithOutput(checkYarnCmd)
		if err != nil {
			return fmt.Errorf("检查包管理器失败: %w", err)
		}

		var installCmd string
		switch pkgManager {
		case "yarn":
			installCmd = "cd {{release_path}} && yarn install --production"
		case "pnpm":
			installCmd = "cd {{release_path}} && pnpm install --production"
		default:
			installCmd = "cd {{release_path}} && npm install --production"
		}

		return exec.RunRemoteCommand(installCmd)
	}

	// 如果无法检测项目类型，跳过安装
	ctx.Logger.Info("无法检测项目类型或没有依赖需要安装，跳过")
	return nil
}
