/*
 * GoDeployer - 灵活、可扩展的部署工具
 * Copyright (C) 2025-2025 amoydavid
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 表示主配置结构
type Config struct {
	Project      string                 `yaml:"project"`
	DefaultStage string                 `yaml:"default_stage"`
	Stages       map[string]StageConfig `yaml:"stages"`
	Options      OptionsConfig          `yaml:"options"`
	Tasks        map[string]TaskConfig  `yaml:"tasks"`
	Hooks        map[string][]string    `yaml:"hooks"`
	Recipe       []string               `yaml:"recipe"`
	Vars         map[string]interface{} `yaml:"vars"`
}

// StageConfig 表示环境特定配置
type StageConfig struct {
	Server         string                 `yaml:"server"`
	Port           int                    `yaml:"port"`
	PrivateKeyPath string                 `yaml:"private_key_path"`
	RemoteDir      string                 `yaml:"remote_dir"`
	KeepReleases   int                    `yaml:"keep_releases"`
	Vars           map[string]interface{} `yaml:"vars"`
}

// OptionsConfig 表示全局选项
type OptionsConfig struct {
	ReleaseDirFormat string   `yaml:"release_dir_format"`
	SharedDirs       []string `yaml:"shared_dirs"`
	SharedFiles      []string `yaml:"shared_files"`
}

// TaskConfig 表示任务配置
type TaskConfig struct {
	Local   string            `yaml:"local"`
	Remote  string            `yaml:"remote"`
	Upload  UploadConfig      `yaml:"upload"`
	Options map[string]string `yaml:"options"`
}

// UploadConfig 表示上传配置
type UploadConfig struct {
	Source  string `yaml:"source"`
	Dest    string `yaml:"dest"`
	Options string `yaml:"options"`
}

// LoadConfig 从YAML文件加载配置
func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	config := &Config{
		DefaultStage: "production",
		Options: OptionsConfig{
			ReleaseDirFormat: "releases/%Y%m%d%H%M%S",
			SharedDirs:       []string{},
			SharedFiles:      []string{},
		},
		Tasks:  make(map[string]TaskConfig),
		Hooks:  make(map[string][]string),
		Recipe: []string{},
		Vars:   make(map[string]interface{}),
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 为阶段设置默认值
	for name, stage := range config.Stages {
		if stage.KeepReleases <= 0 {
			stage.KeepReleases = 5
		}
		if stage.Port <= 0 {
			stage.Port = 22
		}
		if stage.Vars == nil {
			stage.Vars = make(map[string]interface{})
		}
		config.Stages[name] = stage
	}

	return config, nil
}
