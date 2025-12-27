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
	ctx          *deployctx.DeployContext
	enablePool   bool // 是否启用 SSH 连接池
	currentSocket string // 当前使用的 ControlMaster socket
}

// NewExecutor 创建新的执行器
func NewExecutor(ctx *deployctx.DeployContext) *Executor {
	exec := &Executor{ctx: ctx, enablePool: false}
	// 如果 context 中有 SSH 连接池 socket，自动启用并使用
	if ctx.SSHPoolSocket != "" {
		exec.enablePool = true
		exec.currentSocket = ctx.SSHPoolSocket
	}
	return exec
}

// NewExecutorWithPool 创建新的执行器（启用连接池）
func NewExecutorWithPool(ctx *deployctx.DeployContext) *Executor {
	exec := &Executor{ctx: ctx, enablePool: true}
	exec.initPoolSocket()
	return exec
}

// initPoolSocket 初始化连接池 socket
func (e *Executor) initPoolSocket() {
	if !e.enablePool || e.ctx.DryRun {
		return
	}

	// 如果 context 中已经设置了 SSHPoolSocket，直接使用
	if e.ctx.SSHPoolSocket != "" {
		e.currentSocket = e.ctx.SSHPoolSocket
		return
	}

	// 否则从全局连接池获取
	pool := GetGlobalPool()
	socket, _, err := pool.GetMasterSocket(e.ctx.StageConfig.PrivateKeyPath, e.ctx.StageConfig.Server)
	if err == nil {
		e.currentSocket = socket
	}
}

// SetPoolSocket 设置当前使用的 ControlMaster socket
func (e *Executor) SetPoolSocket(socket string) {
	e.currentSocket = socket
}

// 默认超时时间
const defaultTimeout = 10 * time.Minute

// RunSSHCommand 执行交互式 SSH 命令（用于 SSH 登录）
func RunSSHCommand(args ...string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// BuildSSHArgs 构建 SSH 参数切片（用于 exec.Command）
func BuildSSHArgs(privateKeyPath string) []string {
	args := []string{}
	if privateKeyPath != "" {
		args = append(args, "-i", privateKeyPath)
	}
	return args
}

// BuildSSHCommand 构建完整的 SSH 命令切片
func BuildSSHCommand(privateKeyPath, server, command string) []string {
	args := BuildSSHArgs(privateKeyPath)
	return append(args, server, command)
}

// BuildSSHInlineArgs 构建 SSH 参数字符串（用于 shell 命令中的内联使用）
func BuildSSHInlineArgs(privateKeyPath string) string {
	if privateKeyPath != "" {
		return fmt.Sprintf("-i %s ", privateKeyPath)
	}
	return ""
}

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

	// 构建基础 SSH 参数
	sshArgs := BuildSSHArgs(e.ctx.StageConfig.PrivateKeyPath)

	// 如果启用了连接池，添加 ControlPath
	if e.enablePool && e.currentSocket != "" {
		sshArgs = append(sshArgs, "-S", e.currentSocket)
	}

	sshArgs = append(sshArgs, e.ctx.StageConfig.Server, shellCmd)

	e.ctx.Logger.Infof("执行远程命令: ssh %s", strings.Join(sshArgs, " "))
	cmd := exec.CommandContext(cmdCtx, "ssh", sshArgs...)
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

	// 创建有超时的上下文
	cmdCtx, cancel := context.WithTimeout(e.ctx.Context, defaultTimeout)
	defer cancel()

	// 使用登录 shell 执行命令
	shellCmd := fmt.Sprintf("exec $SHELL -l -c '%s'", resolvedCommand)

	// 构建基础 SSH 参数
	sshArgs := BuildSSHArgs(e.ctx.StageConfig.PrivateKeyPath)

	// 如果启用了连接池，添加 ControlPath
	if e.enablePool && e.currentSocket != "" {
		sshArgs = append(sshArgs, "-S", e.currentSocket)
	}

	sshArgs = append(sshArgs, e.ctx.StageConfig.Server, shellCmd)

	e.ctx.Logger.Infof("执行远程命令: ssh %s", strings.Join(sshArgs, " "))
	cmd := exec.CommandContext(cmdCtx, "ssh", sshArgs...)
	output, err := cmd.CombinedOutput()
	if cmdCtx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("远程命令执行超时(超过 %s): %s", defaultTimeout, resolvedCommand)
	}
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

		// 构建 ssh 参数（在管道命令中以字符串形式）
		sshInlineArgs := BuildSSHInlineArgs(e.ctx.StageConfig.PrivateKeyPath)

		// 根据本地操作系统选择合适的 tar 命令
		tarLocal := "tar -czf - ."
		if runtime.GOOS == "darwin" {
			// 在 macOS 上禁用元数据与扩展属性，并排除垃圾文件
			tarLocal = "COPYFILE_DISABLE=1 tar --no-xattrs --no-mac-metadata --exclude '._*' --exclude '.DS_Store' -czf - ."
		}

		tarCommand := fmt.Sprintf("cd %s && %s | ssh %s%s \"tar -xzf - -C %s\" %s",
			resolvedSource, tarLocal, sshInlineArgs, e.ctx.StageConfig.Server, resolvedDest, optionsStr)

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

		// 使用辅助函数构建 SSH 参数
		keyPart := BuildSSHInlineArgs(e.ctx.StageConfig.PrivateKeyPath)
		scpCommand := fmt.Sprintf("scp %s%s %s %s:%s",
			keyPart,
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
