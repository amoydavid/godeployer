package recipe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	deployctx "deployer/internal/context"
	"deployer/internal/registry"
)

// Recipe 表示部署配方
type Recipe struct {
	registry *registry.TaskRegistry
	tasks    []string
	hooks    map[string][]string
}

// NewRecipe 创建新的配方
func NewRecipe(registry *registry.TaskRegistry, tasks []string, hooks map[string][]string) *Recipe {
	return &Recipe{
		registry: registry,
		tasks:    tasks,
		hooks:    hooks,
	}
}

// Execute 运行配方
func (r *Recipe) Execute(ctx *deployctx.DeployContext) error {
	// 运行全局前置钩子
	if err := r.runHook(ctx, "before_all", nil); err != nil {
		return fmt.Errorf("before 钩子失败: %w", err)
	}

	// 按顺序执行每个任务
	for _, taskName := range r.tasks {
		// 运行任务特定的前置钩子
		hookName := fmt.Sprintf("before_%s", taskName)
		if err := r.runHook(ctx, hookName, nil); err != nil {
			return fmt.Errorf("%s 钩子失败: %w", hookName, err)
		}

		// 获取并执行任务
		task, err := r.registry.Get(taskName)
		if err != nil {
			return fmt.Errorf("任务 '%s' 错误: %w", taskName, err)
		}

		ctx.Logger.Infof("执行任务: %s", taskName)
		ctx.Logger.Infof("描述: %s", task.Description())

		if err := task.Execute(ctx); err != nil {
			// 运行任务特定的失败钩子
			failedTaskHook := fmt.Sprintf("after_%s:failed", taskName)
			failedTaskErr := r.runHook(ctx, failedTaskHook, map[string]interface{}{
				"failed_task": taskName,
				"error":       err.Error(),
			})

			if failedTaskErr != nil {
				// 记录但继续使用原始错误
				ctx.Logger.Errorf("运行 %s 钩子失败: %v", failedTaskHook, failedTaskErr)
			}

			// 运行通用的任务后钩子（无论成功失败）
			afterTaskHook := fmt.Sprintf("after_%s", taskName)
			afterTaskErr := r.runHook(ctx, afterTaskHook, map[string]interface{}{
				"failed_task": taskName,
				"error":       err.Error(),
				"status":      "failed",
			})

			if afterTaskErr != nil {
				ctx.Logger.Errorf("运行 %s 钩子失败: %v", afterTaskHook, afterTaskErr)
			}

			// 运行全局失败钩子
			failedErr := r.runHook(ctx, "on_failed", map[string]interface{}{
				"failed_task": taskName,
				"error":       err.Error(),
			})

			if failedErr != nil {
				// 记录但继续使用原始错误
				ctx.Logger.Errorf("运行 on_failed 钩子失败: %v", failedErr)
			}

			return fmt.Errorf("任务 '%s' 失败: %w", taskName, err)
		}

		// 运行任务特定的成功钩子
		successHook := fmt.Sprintf("after_%s:success", taskName)
		if err := r.runHook(ctx, successHook, map[string]interface{}{
			"task":   taskName,
			"status": "success",
		}); err != nil {
			ctx.Logger.Errorf("运行 %s 钩子失败: %v", successHook, err)
		}

		// 运行任务特定的后置钩子（无论成功失败）
		afterHook := fmt.Sprintf("after_%s", taskName)
		if err := r.runHook(ctx, afterHook, map[string]interface{}{
			"task":   taskName,
			"status": "success",
		}); err != nil {
			return fmt.Errorf("%s 钩子失败: %w", afterHook, err)
		}
	}

	// 运行全局后置钩子
	if err := r.runHook(ctx, "after_all", nil); err != nil {
		return fmt.Errorf("after 钩子失败: %w", err)
	}

	return nil
}

// runHook 执行指定名称的钩子（如果存在）
func (r *Recipe) runHook(ctx *deployctx.DeployContext, name string, extraVars map[string]interface{}) error {
	hooks, exists := r.hooks[name]
	if !exists || len(hooks) == 0 {
		return nil // 没有钩子要运行
	}

	// 向上下文添加额外变量
	if extraVars != nil {
		for k, v := range extraVars {
			ctx.AddVar(k, v)
		}
	}

	ctx.Logger.Infof("运行钩子: %s", name)

	for _, hook := range hooks {
		// 尝试查找同名的已注册任务
		task, err := r.registry.Get(hook)
		if err == nil {
			// 如果是已注册任务，执行它
			if err := task.Execute(ctx); err != nil {
				return fmt.Errorf("钩子任务 '%s' 失败: %w", hook, err)
			}
		} else {
			// 否则，将其视为命令
			executor := NewHookExecutor(ctx)
			if err := executor.ExecuteHook(hook); err != nil {
				return fmt.Errorf("钩子命令 '%s' 失败: %w", hook, err)
			}
		}
	}

	return nil
}

// HookExecutor 处理钩子执行
type HookExecutor struct {
	ctx *deployctx.DeployContext
}

// NewHookExecutor 创建新的钩子执行器
func NewHookExecutor(ctx *deployctx.DeployContext) *HookExecutor {
	return &HookExecutor{ctx: ctx}
}

// ExecuteHook 运行钩子命令
func (e *HookExecutor) ExecuteHook(command string) error {
	// 解析命令中的变量
	resolvedCommand := e.ctx.ResolveVar(command)

	// 检查是否是远程命令
	if strings.HasPrefix(resolvedCommand, "remote:") {
		cmd := strings.TrimPrefix(resolvedCommand, "remote:")
		cmd = strings.TrimSpace(cmd)

		e.ctx.Logger.Infof("运行远程钩子: %s", cmd)
		if e.ctx.DryRun {
			return nil
		}

		// 在远程服务器上执行
		return e.executeRemote(cmd)
	}

	// 否则，它是本地命令
	e.ctx.Logger.Infof("运行本地钩子: %s", resolvedCommand)
	if e.ctx.DryRun {
		return nil
	}

	// 在本地执行
	return e.executeLocal(resolvedCommand)
}

// executeLocal 在本地运行命令
func (e *HookExecutor) executeLocal(command string) error {
	// 创建有超时的上下文
	cmdCtx, cancel := context.WithTimeout(e.ctx.Context, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if cmdCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("钩子执行超时(超过5分钟): %s", command)
	}
	return err
}

// executeRemote 在远程服务器上运行命令
func (e *HookExecutor) executeRemote(command string) error {
	// 创建有超时的上下文
	cmdCtx, cancel := context.WithTimeout(e.ctx.Context, 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "ssh", e.ctx.StageConfig.Server, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if cmdCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("远程钩子执行超时(超过5分钟): %s", command)
	}
	return err
}
