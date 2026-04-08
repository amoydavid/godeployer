package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"al.essio.dev/pkg/shellescape"
	"deployer/internal/config"
	"deployer/pkg/logger"
)

// DeployContext 保存部署的状态和配置
type DeployContext struct {
	Context       context.Context
	Config        *config.Config
	Stage         string
	StageConfig   config.StageConfig
	ReleasePath   string
	Vars          map[string]interface{}
	Logger        *logger.Logger
	DryRun        bool
	SSHPoolSocket string // SSH 连接池 ControlMaster socket 路径
}

// NewDeployContext 创建新的部署上下文
func NewDeployContext(ctx context.Context, cfg *config.Config, stage string, dryRun bool) (*DeployContext, error) {
	stageConfig, ok := cfg.Stages[stage]
	if !ok {
		return nil, fmt.Errorf("阶段 '%s' 未定义", stage)
	}

	// 验证私钥文件
	if err := validatePrivateKey(stageConfig.PrivateKeyPath); err != nil {
		return nil, fmt.Errorf("私钥验证失败: %w", err)
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

// ResolveVar 解析字符串中的变量引用，并对变量值进行安全转义以防止命令注入
func (c *DeployContext) ResolveVar(input string) string {
	result := input

	// 替换 {{var}} 模式
	for name, value := range c.Vars {
		placeholder := fmt.Sprintf("{{%s}}", name)

		// 使用 shellescape.Quote 转义特殊字符，防止命令注入
		// Quote 会添加单引号包裹，这是在shell命令中安全使用变量的标准方式
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = shellescape.Quote(v)
		default:
			strValue = shellescape.Quote(fmt.Sprintf("%v", v))
		}

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

// validatePrivateKey 验证私钥文件是否存在且权限正确
func validatePrivateKey(path string) error {
	if path == "" {
		// 未配置私钥路径，这是正常的（可能使用密码认证）
		return nil
	}

	// 检查文件是否存在
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("私钥文件不存在: %s", path)
		}
		return fmt.Errorf("无法访问私钥文件: %w", err)
	}

	// 检查是否为常规文件
	if info.IsDir() {
		return fmt.Errorf("私钥路径指向的是目录而非文件: %s", path)
	}

	// 检查文件权限（应该限制为所有者可读）
	// Unix/Linux: 应该是 600 或更严格
	// Windows: 权限模型不同，只检查文件是否存在
	perm := info.Mode().Perm()

	// 检查是否对其他用户可读 (others readable)
	if perm&0004 != 0 {
		return fmt.Errorf("私钥文件权限过于宽松: %s\n当前权限: %s\n建议执行: chmod 600 %s",
			path, perm, path)
	}

	// 检查是否对组可读 (group readable)
	if perm&0040 != 0 {
		return fmt.Errorf("私钥文件权限过于宽松: %s\n当前权限: %s\n建议执行: chmod 600 %s",
			path, perm, path)
	}

	return nil
}
