package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 结构体定义了部署器所需的所有配置项
type Config struct {
	Host         string   `yaml:"host"`
	SSHUser      string   `yaml:"ssh_user"`
	SSHPort      string   `yaml:"ssh_port"`
	SSHKeyPath   string   `yaml:"ssh_key_path"`
	DeployPath   string   `yaml:"deploy_path"`
	Repository   string   `yaml:"repository"`
	Branch       string   `yaml:"branch"`
	Releases     int      `yaml:"keep_releases"`
	SharedFiles  []string `yaml:"shared_files"`
	SharedDirs   []string `yaml:"shared_dirs"`
	BeforeDeploy []string `yaml:"before_deploy"`
	AfterDeploy  []string `yaml:"after_deploy"`
	// 新增字段
	RemotePath     string `yaml:"remote_path"`
	RestartCommand string `yaml:"restart_command"`
}

// LoadConfig 从指定的YAML文件加载配置
func LoadConfig(filename string) (*Config, error) {
	// 如果文件名不是绝对路径，则相对于当前工作目录
	if !filepath.IsAbs(filename) {
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %v", err)
		}
		filename = filepath.Join(currentDir, filename)
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// 设置默认值
	if config.SSHPort == "" {
		config.SSHPort = "22"
	}
	if config.Branch == "" {
		config.Branch = "main"
	}
	if config.Releases == 0 {
		config.Releases = 5
	}

	// 验证必要的字段
	if config.Host == "" {
		return nil, fmt.Errorf("host is required in config file")
	}
	if config.SSHUser == "" {
		return nil, fmt.Errorf("ssh_user is required in config file")
	}
	if config.SSHKeyPath == "" {
		return nil, fmt.Errorf("ssh_key_path is required in config file")
	}
	if config.DeployPath == "" {
		return nil, fmt.Errorf("deploy_path is required in config file")
	}
	if config.Repository == "" {
		return nil, fmt.Errorf("repository is required in config file")
	}

	// 展开 SSH 密钥路径中的 ~ 为用户主目录
	if config.SSHKeyPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		config.SSHKeyPath = filepath.Join(home, config.SSHKeyPath[2:])
	}

	return &config, nil
}

// GetString 返回指定键的字符串值
func (c *Config) GetString(key string) string {
	switch key {
	case "host":
		return c.Host
	case "ssh_user":
		return c.SSHUser
	case "ssh_port":
		return c.SSHPort
	case "ssh_key_path":
		return c.SSHKeyPath
	case "deploy_path":
		return c.DeployPath
	case "repository":
		return c.Repository
	case "branch":
		return c.Branch
	default:
		return ""
	}
}

// GetInt 返回指定键的整数值
func (c *Config) GetInt(key string) int {
	switch key {
	case "keep_releases":
		return c.Releases
	default:
		return 0
	}
}

// GetStringSlice 返回指定键的字符串切片
func (c *Config) GetStringSlice(key string) []string {
	switch key {
	case "shared_files":
		return c.SharedFiles
	case "shared_dirs":
		return c.SharedDirs
	case "before_deploy":
		return c.BeforeDeploy
	case "after_deploy":
		return c.AfterDeploy
	default:
		return nil
	}
}
