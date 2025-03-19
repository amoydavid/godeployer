package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"deployer/internal/config"
	"deployer/pkg/logger"
)

// DeployContext 保存部署的状态和配置
type DeployContext struct {
	Context     context.Context
	Config      *config.Config
	Stage       string
	StageConfig config.StageConfig
	ReleasePath string
	Vars        map[string]interface{}
	Logger      *logger.Logger
	DryRun      bool
}

// NewDeployContext 创建新的部署上下文
func NewDeployContext(ctx context.Context, cfg *config.Config, stage string, dryRun bool) (*DeployContext, error) {
	stageConfig, ok := cfg.Stages[stage]
	if !ok {
		return nil, fmt.Errorf("阶段 '%s' 未定义", stage)
	}

	// 生成发布路径
	timestamp := time.Now().Format("20060102150405")
	releaseDirFormat := cfg.Options.ReleaseDirFormat
	if releaseDirFormat == "" {
		releaseDirFormat = "releases/%Y%m%d%H%M%S"
	}

	// 替换格式字符串
	releaseDirFormat = strings.ReplaceAll(releaseDirFormat, "%Y", timestamp[0:4])
	releaseDirFormat = strings.ReplaceAll(releaseDirFormat, "%m", timestamp[4:6])
	releaseDirFormat = strings.ReplaceAll(releaseDirFormat, "%d", timestamp[6:8])
	releaseDirFormat = strings.ReplaceAll(releaseDirFormat, "%H", timestamp[8:10])
	releaseDirFormat = strings.ReplaceAll(releaseDirFormat, "%M", timestamp[10:12])
	releaseDirFormat = strings.ReplaceAll(releaseDirFormat, "%S", timestamp[12:14])

	releasePath := filepath.Join(stageConfig.RemoteDir, releaseDirFormat)

	// 初始化变量
	vars := map[string]interface{}{
		"project":       cfg.Project,
		"stage":         stage,
		"remote_dir":    stageConfig.RemoteDir,
		"release_path":  releasePath,
		"timestamp":     timestamp,
		"keep_releases": stageConfig.KeepReleases,
	}

	// 添加自定义变量
	for k, v := range cfg.Vars {
		vars[k] = v
	}

	// 添加阶段特定变量
	for k, v := range stageConfig.Vars {
		vars[k] = v
	}

	return &DeployContext{
		Context:     ctx,
		Config:      cfg,
		Stage:       stage,
		StageConfig: stageConfig,
		ReleasePath: releasePath,
		Vars:        vars,
		Logger:      logger.NewLogger(os.Stdout, !dryRun),
		DryRun:      dryRun,
	}, nil
}

// ResolveVar 解析字符串中的变量引用
func (c *DeployContext) ResolveVar(input string) string {
	result := input

	// 替换 {{var}} 模式
	for name, value := range c.Vars {
		placeholder := fmt.Sprintf("{{%s}}", name)
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result
}

// AddVar 向上下文添加变量
func (c *DeployContext) AddVar(name string, value interface{}) {
	c.Vars[name] = value
}

// ResolveVarsInMap 解析映射中所有字符串值中的变量
func (c *DeployContext) ResolveVarsInMap(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range input {
		if str, ok := v.(string); ok {
			result[k] = c.ResolveVar(str)
		} else {
			result[k] = v
		}
	}

	return result
}
