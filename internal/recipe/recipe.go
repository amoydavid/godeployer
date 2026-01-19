package recipe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	deployctx "deployer/internal/context"
	"deployer/internal/executor"
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
	// 初始化 SSH 连接池（提升性能）
	pool := executor.GetGlobalPool()
	defer pool.Close() // 确保部署结束时关闭所有连接

	// 为当前服务器创建 ControlMaster 连接
	if !ctx.DryRun {
		socket, isNew, err := pool.GetMasterSocket(ctx.StageConfig.PrivateKeyPath, ctx.StageConfig.Server, ctx.StageConfig.Port)
		if err != nil {
			ctx.Logger.Warnf("Failed to create SSH connection pool: %v, using direct connection", err)
		} else {
			ctx.SSHPoolSocket = socket // 将 socket 路径存储到 context 中
			if isNew {
				ctx.Logger.Infof("Created SSH connection pool: %s", socket)
			}
		}
	}

	// 运行全局前置钩子
	if err := r.runBeforeAll(ctx); err != nil {
		return err
	}

	// 执行所有任务
	taskErr := r.executeTasks(ctx)

	// 运行全局后置钩子
	if err := r.runAfterAll(ctx); err != nil {
		return err
	}

	return taskErr
}

// executeTasks 执行所有任务
func (r *Recipe) executeTasks(ctx *deployctx.DeployContext) error {
	for i, taskName := range r.tasks {
		current := i + 1
		total := len(r.tasks)

		if err := r.executeSingleTask(ctx, taskName, current, total); err != nil {
			return err
		}
	}
	return nil
}

// executeSingleTask 执行单个任务及其相关钩子
func (r *Recipe) executeSingleTask(ctx *deployctx.DeployContext, taskName string, current, total int) error {
	// 运行前置钩子
	if err := r.runBeforeTask(ctx, taskName); err != nil {
		return err
	}

	// 执行任务本身
	taskErr := r.runTask(ctx, taskName, current, total)

	// 运行后置钩子
	if err := r.runAfterTask(ctx, taskName, taskErr); err != nil {
		return err
	}

	return taskErr
}

// runTask 执行单个任务
func (r *Recipe) runTask(ctx *deployctx.DeployContext, taskName string, current, total int) error {
	// 获取任务
	task, err := r.registry.Get(taskName)
	if err != nil {
		return fmt.Errorf("任务 '%s' 错误: %w", taskName, err)
	}

	// 显示任务信息
	ctx.Logger.Infof("[%d/%d] Executing task: %s", current, total, taskName)
	ctx.Logger.Infof("Description: %s", task.Description())

	// 执行任务
	return task.Execute(ctx)
}

// runBeforeAll 运行全局前置钩子
func (r *Recipe) runBeforeAll(ctx *deployctx.DeployContext) error {
	beforeHooks := ctx.Config.Hooks["before_deploy"]
	if len(beforeHooks) == 0 {
		return nil
	}

	ctx.Logger.Info("Running pre-deploy hooks...")
	for _, hook := range beforeHooks {
		if err := r.runHook(ctx, hook, nil); err != nil {
			return err
		}
	}
	return nil
}

// runAfterAll 运行全局后置钩子
func (r *Recipe) runAfterAll(ctx *deployctx.DeployContext) error {
	afterHooks := ctx.Config.Hooks["after_deploy"]
	if len(afterHooks) == 0 {
		return nil
	}

	ctx.Logger.Info("Running post-deploy hooks...")
	for _, hook := range afterHooks {
		if err := r.runHook(ctx, hook, nil); err != nil {
			return err
		}
	}
	return nil
}

// runBeforeTask 运行任务前置钩子
func (r *Recipe) runBeforeTask(ctx *deployctx.DeployContext, taskName string) error {
	hookName := fmt.Sprintf("before_%s", taskName)
	if hooks, exists := r.hooks[hookName]; exists && len(hooks) > 0 {
		ctx.Logger.Infof("Running %s hooks...", hookName)
		for _, hook := range hooks {
			if err := r.runHook(ctx, hook, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

// runAfterTask 运行任务后置钩子
func (r *Recipe) runAfterTask(ctx *deployctx.DeployContext, taskName string, taskErr error) error {
	if taskErr != nil {
		// 任务失败，运行失败钩子
		return r.runTaskFailedHooks(ctx, taskName, taskErr)
	}

	// 任务成功，运行成功钩子
	return r.runTaskSuccessHooks(ctx, taskName)
}

// runTaskSuccessHooks 运行任务成功后的钩子
func (r *Recipe) runTaskSuccessHooks(ctx *deployctx.DeployContext, taskName string) error {
	// 运行成功特定钩子
	successHook := fmt.Sprintf("after_%s:success", taskName)
	if hooks, exists := r.hooks[successHook]; exists {
		for _, hook := range hooks {
			if err := r.runHook(ctx, hook, map[string]interface{}{
				"task":   taskName,
				"status": "success",
			}); err != nil {
				ctx.Logger.Errorf("Failed to run %s hook: %v", successHook, err)
			}
		}
	}

	// 运行通用后置钩子
	afterHook := fmt.Sprintf("after_%s", taskName)
	if hooks, exists := r.hooks[afterHook]; exists {
		for _, hook := range hooks {
			if err := r.runHook(ctx, hook, map[string]interface{}{
				"task":   taskName,
				"status": "success",
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// runTaskFailedHooks 运行任务失败后的钩子
func (r *Recipe) runTaskFailedHooks(ctx *deployctx.DeployContext, taskName string, taskErr error) error {
	// 运行失败特定钩子
	failedTaskHook := fmt.Sprintf("after_%s:failed", taskName)
	if hooks, exists := r.hooks[failedTaskHook]; exists {
		for _, hook := range hooks {
			if err := r.runHook(ctx, hook, map[string]interface{}{
				"failed_task": taskName,
				"error":       taskErr.Error(),
			}); err != nil {
				ctx.Logger.Errorf("Failed to run %s hook: %v", failedTaskHook, err)
			}
		}
	}

	// 运行通用后置钩子
	afterHook := fmt.Sprintf("after_%s", taskName)
	if hooks, exists := r.hooks[afterHook]; exists {
		for _, hook := range hooks {
			if err := r.runHook(ctx, hook, map[string]interface{}{
				"failed_task": taskName,
				"error":       taskErr.Error(),
				"status":      "failed",
			}); err != nil {
				ctx.Logger.Errorf("Failed to run %s hook: %v", afterHook, err)
			}
		}
	}

	// 运行全局失败钩子
	if hooks, exists := r.hooks["on_failed"]; exists {
		for _, hook := range hooks {
			if err := r.runHook(ctx, hook, map[string]interface{}{
				"failed_task": taskName,
				"error":       taskErr.Error(),
			}); err != nil {
				ctx.Logger.Errorf("Failed to run on_failed hook: %v", err)
			}
		}
	}

	return fmt.Errorf("任务 '%s' 失败: %w", taskName, taskErr)
}

// runHook 执行单个钩子
func (r *Recipe) runHook(ctx *deployctx.DeployContext, hook string, extraVars map[string]interface{}) error {
	// 向上下文添加额外变量
	if extraVars != nil {
		for k, v := range extraVars {
			ctx.AddVar(k, v)
		}
	}

	// 尝试查找同名的已注册任务
	task, err := r.registry.Get(hook)
	if err == nil {
		// 如果是已注册任务，执行它
		ctx.Logger.Infof("Executing hook task: %s", hook)
		if err := task.Execute(ctx); err != nil {
			return fmt.Errorf("钩子任务 '%s' 失败: %w", hook, err)
		}
		return nil
	}

	// 否则，将其视为命令
	executor := NewHookExecutor(ctx)
	if err := executor.ExecuteHook(hook); err != nil {
		return fmt.Errorf("钩子命令 '%s' 失败: %w", hook, err)
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

		e.ctx.Logger.Infof("Running remote hook: %s", cmd)
		if e.ctx.DryRun {
			return nil
		}

		// 在远程服务器上执行
		return e.executeRemote(cmd)
	}

	// 否则，它是本地命令
	e.ctx.Logger.Infof("Running local hook: %s", resolvedCommand)
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

	// 使用辅助函数构建 SSH 参数
	sshArgs := executor.BuildSSHCommand(e.ctx.StageConfig.PrivateKeyPath, e.ctx.StageConfig.Port, e.ctx.StageConfig.Server, command)
	e.ctx.Logger.Infof("Executing remote command: ssh %s", strings.Join(sshArgs, " "))
	cmd := exec.CommandContext(cmdCtx, "ssh", sshArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if cmdCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("远程钩子执行超时(超过5分钟): %s", command)
	}
	return err
}
