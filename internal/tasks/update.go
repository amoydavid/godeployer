package tasks

import (
	"fmt"
	"os"

	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// UpdateCodeTask 表示更新代码的任务
type UpdateCodeTask struct{}

// NewUpdateCodeTask 创建新的更新代码任务
func NewUpdateCodeTask() task.Task {
	return &UpdateCodeTask{}
}

// Name 返回任务名称
func (t *UpdateCodeTask) Name() string {
	return "update_code"
}

// Description 返回任务描述
func (t *UpdateCodeTask) Description() string {
	return "准备并上传代码到远程服务器"
}

// Execute 执行更新代码任务
func (t *UpdateCodeTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)
	taskCfg, ok := ctx.Config.Tasks["update_code"]

	// 如果配置中有上传配置，使用它
	if ok && taskCfg.Upload.Source != "" && taskCfg.Upload.Dest != "" {
		// 如果有需要先执行的本地命令
		if taskCfg.Local != "" {
			if err := exec.RunLocalCommand(taskCfg.Local); err != nil {
				return fmt.Errorf("本地命令执行失败: %w", err)
			}
		}

		// 执行上传
		dest := taskCfg.Upload.Dest
		if dest == "" {
			dest = "{{release_path}}"
		}

		options := taskCfg.Upload.Options
		return exec.UploadDirectory(taskCfg.Upload.Source, dest, options)
	}

	// 否则，使用默认上传逻辑
	tempDir := ".deploy_tmp"

	// 创建临时目录
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir) // 清理临时目录

	// 复制当前目录内容到临时目录
	// 忽略 .git, node_modules 等目录
	copyCmd := fmt.Sprintf(`find . -type f -not -path "*/\.*" -not -path "*/node_modules/*" -not -path "*/%s/*" | xargs -I{} cp --parents {} %s/`, tempDir, tempDir)
	if err := exec.RunLocalCommand(copyCmd); err != nil {
		return fmt.Errorf("复制文件到临时目录失败: %w", err)
	}

	// 上传到远程服务器
	return exec.UploadDirectory(tempDir, "{{release_path}}", "")
}
