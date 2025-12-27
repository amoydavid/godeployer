package tasks

import (
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
	"fmt"
	"strings"
)

// RollbackTask 表示回滚到之前发布版本的任务
type RollbackTask struct {
	Steps int // 回滚的步数，默认为1
}

// NewRollbackTask 创建新的回滚任务
func NewRollbackTask(steps int) task.Task {
	if steps <= 0 {
		steps = 1
	}
	return &RollbackTask{Steps: steps}
}

// SetSteps 设置回滚的步数
func (t *RollbackTask) SetSteps(steps int) {
	if steps <= 0 {
		steps = 1
	}
	t.Steps = steps
}

// Name 返回任务名称
func (t *RollbackTask) Name() string {
	return "rollback"
}

// Description 返回任务描述
func (t *RollbackTask) Description() string {
	return fmt.Sprintf("回滚到之前的%d个发布版本", t.Steps)
}

// Execute 执行回滚任务
func (t *RollbackTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 如果配置中有自定义命令，使用它
	cmd, ok := ctx.GetTaskRemoteCmd("rollback")
	if ok {
		return exec.RunRemoteCommand(cmd)
	}

	// 否则，执行默认回滚逻辑

	// 1. 获取当前版本的信息
	getCurrentVersionCmd := "ls -l {{remote_dir}}/current | awk '{print $NF}'"
	currentVersionOutput, err := exec.CaptureRemoteCommandOutput(getCurrentVersionCmd)
	if err != nil {
		return fmt.Errorf("获取当前版本失败: %w", err)
	}

	ctx.Logger.Info(fmt.Sprintf("当前版本: %s", currentVersionOutput))

	// 2. 获取可用的历史版本列表
	listVersionsCmd := "ls -dt {{remote_dir}}/releases/* | head -n " + fmt.Sprintf("%d", t.Steps+1)
	versionsOutput, err := exec.CaptureRemoteCommandOutput(listVersionsCmd)
	if err != nil {
		return fmt.Errorf("获取历史版本列表失败: %w", err)
	}

	if versionsOutput == "" {
		return fmt.Errorf("没有找到可用的发布版本")
	}

	versions := strings.Split(strings.TrimSpace(versionsOutput), "\n")
	if len(versions) <= 1 {
		return fmt.Errorf("没有足够的历史版本用于回滚，找到的版本数: %d", len(versions))
	}

	// 3. 根据回滚步数选择目标版本
	// 添加边界检查，防止数组越界
	if t.Steps < 0 || t.Steps >= len(versions) {
		return fmt.Errorf("回滚步数 %d 超出可用版本范围 [0, %d)，当前共有 %d 个版本",
			t.Steps, len(versions), len(versions))
	}

	targetVersion := versions[t.Steps]
	ctx.Logger.Info(fmt.Sprintf("回滚到版本: %s", targetVersion))

	// 4. 更新 current 链接到目标版本
	rollbackCmd := "ln -sfn " + targetVersion + " {{remote_dir}}/current"
	if err := exec.RunRemoteCommand(rollbackCmd); err != nil {
		return fmt.Errorf("更新当前版本链接失败: %w", err)
	}

	ctx.Logger.Success(fmt.Sprintf("已成功回滚到版本: %s", targetVersion))
	return nil
}
