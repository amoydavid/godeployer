package tasks

import (
	"deployer/internal/registry"
)

// RegisterBuiltinTasks 注册所有内置任务
func RegisterBuiltinTasks(registry *registry.TaskRegistry) {
	// 注册标准任务
	registry.Register(NewBuildTask())
	registry.Register(NewSetupTask())
	registry.Register(NewUpdateCodeTask())
	registry.Register(NewInstallDependenciesTask())
	registry.Register(NewPublishReleaseTask())
	registry.Register(NewSymlinkReleaseTask()) // 注册软链任务
	registry.Register(NewRestartAppTask())
	registry.Register(NewCleanupTask())
	registry.Register(NewRollbackTask(1)) // 添加回滚任务，默认回滚1个版本
}
