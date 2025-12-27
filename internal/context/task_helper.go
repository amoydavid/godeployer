package context

import "deployer/internal/config"

// GetTaskConfig 获取指定任务的配置
func (c *DeployContext) GetTaskConfig(taskName string) (config.TaskConfig, bool) {
	cfg, ok := c.Config.Tasks[taskName]
	return cfg, ok
}

// ShouldRunRemotely 检查任务是否应该在远程执行
func (c *DeployContext) ShouldRunRemotely(taskName string) bool {
	cfg, ok := c.Config.Tasks[taskName]
	return ok && cfg.Remote != ""
}

// GetTaskRemoteCmd 获取任务的远程命令
func (c *DeployContext) GetTaskRemoteCmd(taskName string) (string, bool) {
	cfg, ok := c.Config.Tasks[taskName]
	if !ok {
		return "", false
	}
	return cfg.Remote, cfg.Remote != ""
}

// GetTaskLocalCmd 获取任务的本地命令
func (c *DeployContext) GetTaskLocalCmd(taskName string) (string, bool) {
	cfg, ok := c.Config.Tasks[taskName]
	if !ok {
		return "", false
	}
	return cfg.Local, cfg.Local != ""
}

// HasTaskConfig 检查任务是否有配置
func (c *DeployContext) HasTaskConfig(taskName string) bool {
	_, ok := c.Config.Tasks[taskName]
	return ok
}
