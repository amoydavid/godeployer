package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// SSHPool 管理 SSH 连接池（基于 OpenSSH ControlMaster）
type SSHPool struct {
	mu     sync.Mutex
	sockets map[string]string // server -> socket path
}

// NewSSHPool 创建新的 SSH 连接池
func NewSSHPool() *SSHPool {
	return &SSHPool{
		sockets: make(map[string]string),
	}
}

// GetMasterSocket 获取或创建 ControlMaster socket
// 返回 socket 路径和是否是新创建的连接
func (p *SSHPool) GetMasterSocket(privateKeyPath, server string) (string, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果已有连接，返回 socket 路径
	if socket, exists := p.sockets[server]; exists {
		return socket, false, nil
	}

	// 创建新的 ControlMaster socket
	socketPath, err := p.createMasterSocket(privateKeyPath, server)
	if err != nil {
		return "", false, fmt.Errorf("创建 ControlMaster 失败: %w", err)
	}

	p.sockets[server] = socketPath
	return socketPath, true, nil
}

// createMasterSocket 创建 ControlMaster socket
func (p *SSHPool) createMasterSocket(privateKeyPath, server string) (string, error) {
	// 在临时目录创建 socket
	tmpDir := os.TempDir()
	socketPath := filepath.Join(tmpDir, fmt.Sprintf("ssh-cm-%s.sock", sanitizeServerName(server)))

	// 使用 ControlMaster 创建持久连接
	// ssh -fN -S <socket> -o ControlMaster=yes -o ControlPersist=yes <server>
	args := BuildSSHArgs(privateKeyPath)
	args = append(args,
		"-fN", // 后台执行，不执行远程命令
		"-S", socketPath, // 指定 socket 路径
		"-o", "ControlMaster=yes", // 启用 ControlMaster
		"-o", "ControlPersist=10m", // 连接保持 10 分钟
		server,
	)

	cmd := exec.Command("ssh", args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("启动 ControlMaster 失败: %w", err)
	}

	return socketPath, nil
}

// Close 关闭所有连接
func (p *SSHPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for server, socket := range p.sockets {
		// 使用 ssh -O stop 关闭 ControlMaster
		cmd := exec.Command("ssh", "-S", socket, "-O", "stop", "localhost")
		if err := cmd.Run(); err != nil {
			lastErr = fmt.Errorf("关闭 %s 连接失败: %w", server, err)
		}
		// 同时删除 socket 文件
		os.Remove(socket)
	}

	p.sockets = make(map[string]string)
	return lastErr
}

// sanitizeServerName 将服务器名转换为安全的文件名
func sanitizeServerName(server string) string {
	// 简单替换特殊字符为下划线
	result := ""
	for _, c := range server {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '.' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	return result
}

// 全局连接池实例
var globalPool *SSHPool
var poolOnce sync.Once

// GetGlobalPool 获取全局 SSH 连接池
func GetGlobalPool() *SSHPool {
	poolOnce.Do(func() {
		globalPool = NewSSHPool()
	})
	return globalPool
}
