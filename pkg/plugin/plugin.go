package plugin

import (
	"fmt"
	"path/filepath"
	"plugin"

	"deployer/internal/registry"
)

// Plugin 表示部署工具插件
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Version 返回插件版本
	Version() string

	// Register 向任务注册表注册插件任务
	Register(registry *registry.TaskRegistry) error
}

// LoadPlugin 从.so文件加载插件
func LoadPlugin(path string) (Plugin, error) {
	// 加载插件
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("无法打开插件 %s: %w", path, err)
	}

	// 查找Plugin符号
	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("插件 %s 未导出 'Plugin' 符号: %w", path, err)
	}

	// 断言符号是一个Plugin
	plugin, ok := sym.(Plugin)
	if !ok {
		return nil, fmt.Errorf("插件 %s 未实现 Plugin 接口", path)
	}

	return plugin, nil
}

// LoadPluginsFromDir 从目录加载所有插件
func LoadPluginsFromDir(dir string, registry *registry.TaskRegistry) error {
	// 查找目录中的所有.so文件
	pattern := filepath.Join(dir, "*.so")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("无法列出 %s 中的插件: %w", dir, err)
	}

	// 加载每个插件
	for _, path := range matches {
		plugin, err := LoadPlugin(path)
		if err != nil {
			return fmt.Errorf("无法加载插件 %s: %w", path, err)
		}

		// 注册插件任务
		if err := plugin.Register(registry); err != nil {
			return fmt.Errorf("无法注册插件 %s: %w", plugin.Name(), err)
		}
	}

	return nil
}
