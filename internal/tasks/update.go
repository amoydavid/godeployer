package tasks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	cfg, ok := ctx.GetTaskConfig("update_code")

	// 如果配置中有上传配置，使用它
	if ok && cfg.Upload.Source != "" && cfg.Upload.Dest != "" {
		// 如果有需要先执行的本地命令
		if cfg.Local != "" {
			if err := exec.RunLocalCommand(cfg.Local); err != nil {
				return fmt.Errorf("本地命令执行失败: %w", err)
			}
		}

		// 执行上传
		dest := cfg.Upload.Dest
		if dest == "" {
			dest = "{{release_path}}"
		}

		options := cfg.Upload.Options
		return exec.UploadDirectory(cfg.Upload.Source, dest, options)
	}

	// 否则，使用默认上传逻辑
	// 使用系统临时目录创建唯一的临时文件夹（避免并发冲突）
	tempDir, err := os.MkdirTemp("", "deployer-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir) // 确保清理临时目录

	// 使用 Go 原生方式复制文件到临时目录（跨平台兼容）
	if err := copyFilesToTempDir(tempDir, ctx); err != nil {
		return fmt.Errorf("复制文件到临时目录失败: %w", err)
	}

	// 上传到远程服务器
	return exec.UploadDirectory(tempDir, "{{release_path}}", "")
}

// copyFilesToTempDir 使用 Go 原生方式复制文件到临时目录
func copyFilesToTempDir(tempDir string, ctx *deployctx.DeployContext) error {
	// 需要忽略的目录
	ignoredDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
		"target":       true,
		"build":        true,
		"dist":         true,
	}

	// 需要忽略的文件扩展名
	ignoredExts := map[string]bool{
		".so":    true,
		".exe":   true,
		".dll":   true,
		".dylib": true,
	}

	// 遍历当前目录
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if path == "." {
			return nil
		}

		// 获取路径的各个部分
		pathParts := strings.Split(filepath.ToSlash(path), "/")

		// 跳过隐藏文件和目录
		if len(pathParts) > 0 && strings.HasPrefix(pathParts[0], ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 跳过忽略的目录
		if info.IsDir() {
			if ignoredDirs[pathParts[0]] {
				return filepath.SkipDir
			}
			// 创建目标目录
			destPath := filepath.Join(tempDir, path)
			return os.MkdirAll(destPath, info.Mode())
		}

		// 跳过忽略的文件扩展名
		ext := filepath.Ext(path)
		if ignoredExts[ext] {
			return nil
		}

		// 复制文件
		destPath := filepath.Join(tempDir, path)
		if err := copyFile(path, destPath); err != nil {
			return fmt.Errorf("复制文件 %s 失败: %w", path, err)
		}

		return nil
	})

	return err
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制内容
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
