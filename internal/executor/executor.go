package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	deployctx "deployer/internal/context"
)

// Executor 处理命令执行
type Executor struct {
	ctx *deployctx.DeployContext
}

// NewExecutor 创建新的执行器
func NewExecutor(ctx *deployctx.DeployContext) *Executor {
	return &Executor{ctx: ctx}
}

// 默认超时时间
const defaultTimeout = 10 * time.Minute

// getSystemShell 根据操作系统返回合适的 shell 命令
func getSystemShell() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		return "cmd", []string{"/c"}
	default:
		return "sh", []string{"-c"}
	}
}

// RunLocalCommand 在本地执行命令
func (e *Executor) RunLocalCommand(command string) error {
	resolvedCommand := e.ctx.ResolveVar(command)

	e.ctx.Logger.Infof("执行本地命令: %s", resolvedCommand)
	if e.ctx.DryRun {
		e.ctx.Logger.Infof("[模拟运行] 将执行: %s", resolvedCommand)
		return nil
	}

	// 创建有超时的上下文
	cmdCtx, cancel := context.WithTimeout(e.ctx.Context, defaultTimeout)
	defer cancel()

	shell, args := getSystemShell()
	cmd := exec.CommandContext(cmdCtx, shell, append(args, resolvedCommand)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if cmdCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("命令执行超时(超过 %s): %s", defaultTimeout, resolvedCommand)
	}
	return err
}

// RunLocalCommandWithOutput 在本地执行命令并返回其输出
func (e *Executor) RunLocalCommandWithOutput(command string) (string, error) {
	resolvedCommand := e.ctx.ResolveVar(command)

	e.ctx.Logger.Infof("执行本地命令: %s", resolvedCommand)
	if e.ctx.DryRun {
		e.ctx.Logger.Infof("[模拟运行] 将执行: %s", resolvedCommand)
		return "[模拟运行输出]", nil
	}

	// 创建有超时的上下文
	cmdCtx, cancel := context.WithTimeout(e.ctx.Context, defaultTimeout)
	defer cancel()

	shell, args := getSystemShell()
	cmd := exec.CommandContext(cmdCtx, shell, append(args, resolvedCommand)...)
	output, err := cmd.CombinedOutput()

	if cmdCtx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("命令执行超时(超过 %s): %s", defaultTimeout, resolvedCommand)
	}
	return string(output), err
}

// RunRemoteCommand 在远程服务器上执行命令
func (e *Executor) RunRemoteCommand(command string) error {
	resolvedCommand := e.ctx.ResolveVar(command)

	e.ctx.Logger.Infof("执行远程命令: %s", resolvedCommand)
	if e.ctx.DryRun {
		e.ctx.Logger.Infof("[模拟运行] 将在 %s 上执行: %s",
			e.ctx.StageConfig.Server, resolvedCommand)
		return nil
	}

	// 创建有超时的上下文
	cmdCtx, cancel := context.WithTimeout(e.ctx.Context, defaultTimeout)
	defer cancel()

	// 使用登录 shell 执行命令
	shellCmd := fmt.Sprintf("exec bash -l -c '%s'", resolvedCommand)
	e.ctx.Logger.Infof("执行远程命令: %s %s %s", "ssh", e.ctx.StageConfig.Server, shellCmd)
	cmd := exec.CommandContext(cmdCtx, "ssh", e.ctx.StageConfig.Server, shellCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if cmdCtx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("远程命令执行超时(超过 %s): %s", defaultTimeout, resolvedCommand)
	}
	return err
}

// RunRemoteCommandWithOutput 在远程服务器上执行命令并返回其输出
func (e *Executor) RunRemoteCommandWithOutput(command string) (string, error) {
	resolvedCommand := e.ctx.ResolveVar(command)

	e.ctx.Logger.Infof("执行远程命令: %s", resolvedCommand)
	if e.ctx.DryRun {
		e.ctx.Logger.Infof("[模拟运行] 将在 %s 上执行: %s",
			e.ctx.StageConfig.Server, resolvedCommand)
		return "[模拟运行输出]", nil
	}

	// 使用登录 shell 执行命令
	shellCmd := fmt.Sprintf("exec $SHELL -l -c '%s'", resolvedCommand)
	e.ctx.Logger.Infof("执行远程命令: %s %s %s", "ssh", e.ctx.StageConfig.Server, shellCmd)
	cmd := exec.Command("ssh", e.ctx.StageConfig.Server, shellCmd)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// CaptureRemoteCommandOutput 在远程服务器上执行命令并返回其输出
// 这是 RunRemoteCommandWithOutput 的别名
func (e *Executor) CaptureRemoteCommandOutput(command string) (string, error) {
	return e.RunRemoteCommandWithOutput(command)
}

// UploadDirectory 将本地目录上传到远程服务器
func (e *Executor) UploadDirectory(source, destination string, options string) error {
	resolvedSource := e.ctx.ResolveVar(source)
	resolvedDest := e.ctx.ResolveVar(destination)
	resolvedOptions := e.ctx.ResolveVar(options)

	// 检查源路径是否存在
	sourceInfo, err := os.Stat(resolvedSource)
	if err != nil {
		return fmt.Errorf("源路径不存在: %w", err)
	}

	e.ctx.Logger.Infof("上传从 %s 到 %s:%s",
		resolvedSource, e.ctx.StageConfig.Server, resolvedDest)

	if e.ctx.DryRun {
		e.ctx.Logger.Infof("[模拟运行] 将上传 %s 到 %s:%s",
			resolvedSource, e.ctx.StageConfig.Server, resolvedDest)
		return nil
	}

	// 首先创建远程目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", resolvedDest)
	if err := e.RunRemoteCommand(mkdirCmd); err != nil {
		return fmt.Errorf("创建远程目录失败: %w", err)
	}

	// 根据源路径类型选择上传方式
	if sourceInfo.IsDir() {
		// 如果是目录，使用 tar 上传
		optionsStr := ""
		if resolvedOptions != "" {
			optionsStr = resolvedOptions
		}

		tarCommand := fmt.Sprintf("cd %s && tar -czf - . | ssh %s \"tar -xzf - -C %s\" %s",
			resolvedSource, e.ctx.StageConfig.Server, resolvedDest, optionsStr)

		cmd := exec.Command("sh", "-c", tarCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	} else {
		// 如果是单个文件，使用 scp 上传
		optionsStr := ""
		if resolvedOptions != "" {
			optionsStr = resolvedOptions
		}

		// 构建目标路径：如果目标是目录，则保持原文件名
		targetPath := resolvedDest
		if strings.HasSuffix(resolvedDest, "/") {
			targetPath = resolvedDest + sourceInfo.Name()
		}

		scpCommand := fmt.Sprintf("scp %s %s %s:%s",
			optionsStr,
			resolvedSource,
			e.ctx.StageConfig.Server,
			targetPath)

		cmd := exec.Command("sh", "-c", scpCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}
}
