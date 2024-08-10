package tasks

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/amoydavid/godeployer/internal/config"
	"github.com/amoydavid/godeployer/internal/ssh"
)

func Rollback(client *ssh.Client, cfg *config.Config, version string) error {
	if version == "" {
		return rollbackToPrevious(client, cfg)
	} else {
		return rollbackToVersion(client, cfg, version)
	}
}

func rollbackToPrevious(client *ssh.Client, cfg *config.Config) error {
	// 获取所有版本目录
	versions, err := listVersions(client, cfg)
	if err != nil {
		return fmt.Errorf("failed to list versions: %v", err)
	}

	if len(versions) < 2 {
		return fmt.Errorf("not enough versions to rollback")
	}

	// 获取当前版本和上一个版本
	currentVersion := versions[len(versions)-1]
	previousVersion := versions[len(versions)-2]

	// 执行回滚
	return performRollback(client, cfg, previousVersion, currentVersion)
}

func rollbackToVersion(client *ssh.Client, cfg *config.Config, version string) error {
	// 检查指定版本是否存在
	versions, err := listVersions(client, cfg)
	if err != nil {
		return fmt.Errorf("failed to list versions: %v", err)
	}

	versionExists := false
	for _, v := range versions {
		if v == version {
			versionExists = true
			break
		}
	}

	if !versionExists {
		return fmt.Errorf("specified version %s does not exist", version)
	}

	// 获取当前版本
	currentVersion := versions[len(versions)-1]

	// 执行回滚
	return performRollback(client, cfg, version, currentVersion)
}

func listVersions(client *ssh.Client, cfg *config.Config) ([]string, error) {
	cmd := fmt.Sprintf("ls -1 %s", cfg.RemotePath)
	output, err := client.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %v", err)
	}

	versions := strings.Split(strings.TrimSpace(output), "\n")
	sort.Strings(versions)
	return versions, nil
}

func performRollback(client *ssh.Client, cfg *config.Config, targetVersion, currentVersion string) error {
	// 更新符号链接
	currentLink := filepath.Join(cfg.RemotePath, "current")
	targetPath := filepath.Join(cfg.RemotePath, targetVersion)
	cmd := fmt.Sprintf("ln -sfn %s %s", targetPath, currentLink)
	_, err := client.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to update symlink: %v", err)
	}

	// 重启应用（如果需要）
	if cfg.RestartCommand != "" {
		_, err = client.RunCommand(cfg.RestartCommand)
		if err != nil {
			return fmt.Errorf("failed to restart application: %v", err)
		}
	}

	// 重命名当前版本目录为回滚时间戳
	timestamp := time.Now().Format("20060102_150405")
	rollbackDir := filepath.Join(cfg.RemotePath, fmt.Sprintf("%s_rollback_%s", currentVersion, timestamp))
	cmd = fmt.Sprintf("mv %s %s", filepath.Join(cfg.RemotePath, currentVersion), rollbackDir)
	_, err = client.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to rename current version directory: %v", err)
	}

	return nil
}
