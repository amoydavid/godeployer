package tasks

import (
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/task"
)

// PublishReleaseTask 表示发布发布的任务
type PublishReleaseTask struct{}

// NewPublishReleaseTask 创建新的发布任务
func NewPublishReleaseTask() task.Task {
	return &PublishReleaseTask{}
}

// Name 返回任务名称
func (t *PublishReleaseTask) Name() string {
	return "publish_release"
}

// Description 返回任务描述
func (t *PublishReleaseTask) Description() string {
	return "设置共享目录和文件"
}

// Execute 执行发布任务
func (t *PublishReleaseTask) Execute(ctx *deployctx.DeployContext) error {
	exec := executor.NewExecutor(ctx)

	// 如果配置中有自定义命令，使用它
	cmd, ok := ctx.GetTaskRemoteCmd("publish_release")
	if ok {
		return exec.RunRemoteCommand(cmd)
	}

	// 设置共享目录
	if err := t.setupSharedDirs(ctx, exec); err != nil {
		return err
	}

	// 设置共享文件
	if err := t.setupSharedFiles(ctx, exec); err != nil {
		return err
	}

	ctx.Logger.Success("Release configuration successful, shared directories and files have been set up")
	return nil
}

// setupSharedDirs 设置所有共享目录
func (t *PublishReleaseTask) setupSharedDirs(ctx *deployctx.DeployContext, exec *executor.Executor) error {
	for _, dir := range ctx.Config.Options.SharedDirs {
		if err := t.setupSharedDir(exec, dir); err != nil {
			return err
		}
	}
	return nil
}

// setupSharedDir 设置单个共享目录
func (t *PublishReleaseTask) setupSharedDir(exec *executor.Executor, dir string) error {
	// 确保共享目录存在
	createSharedDirCmd := "mkdir -p {{remote_dir}}/shared/" + dir
	if err := exec.RunRemoteCommand(createSharedDirCmd); err != nil {
		return err
	}

	// 如果共享目录是空的，但release目录下有对应目录，则移动内容
	moveIfEmptyCmd := `
if [ -d {{release_path}}/` + dir + ` ] && [ ! -L {{release_path}}/` + dir + ` ] && [ ! "$(ls -A {{remote_dir}}/shared/` + dir + `)" ]; then
    mv {{release_path}}/` + dir + `/* {{remote_dir}}/shared/` + dir + `/ 2>/dev/null || true
fi
`
	if err := exec.RunRemoteCommand(moveIfEmptyCmd); err != nil {
		return err
	}

	// 删除发布版本中的目录
	removeReleaseDir := "rm -rf {{release_path}}/" + dir
	if err := exec.RunRemoteCommand(removeReleaseDir); err != nil {
		return err
	}

	// 创建从共享目录到发布版本的软链接
	linkDirCmd := "ln -sfn {{remote_dir}}/shared/" + dir + " {{release_path}}/" + dir
	return exec.RunRemoteCommand(linkDirCmd)
}

// setupSharedFiles 设置所有共享文件
func (t *PublishReleaseTask) setupSharedFiles(ctx *deployctx.DeployContext, exec *executor.Executor) error {
	for _, file := range ctx.Config.Options.SharedFiles {
		if err := t.setupSharedFile(exec, file); err != nil {
			return err
		}
	}
	return nil
}

// setupSharedFile 设置单个共享文件
func (t *PublishReleaseTask) setupSharedFile(exec *executor.Executor, file string) error {
	// 确保共享文件的目录存在
	createSharedFileDir := "mkdir -p $(dirname {{remote_dir}}/shared/" + file + ")"
	if err := exec.RunRemoteCommand(createSharedFileDir); err != nil {
		return err
	}

	// 如果共享文件不存在，但发布中有，将其移动到共享位置
	moveIfExistsCmd := `
if [ ! -f {{remote_dir}}/shared/` + file + ` ] && [ -f {{release_path}}/` + file + ` ] && [ ! -L {{release_path}}/` + file + ` ]; then
    mv {{release_path}}/` + file + ` {{remote_dir}}/shared/` + file + `
fi
`
	if err := exec.RunRemoteCommand(moveIfExistsCmd); err != nil {
		return err
	}

	// 删除发布版本中的文件(如果还存在)
	removeReleaseFile := "rm -f {{release_path}}/" + file
	if err := exec.RunRemoteCommand(removeReleaseFile); err != nil {
		return err
	}

	// 创建从共享文件到发布版本的软链接
	linkFileCmd := "ln -sf {{remote_dir}}/shared/" + file + " {{release_path}}/" + file
	return exec.RunRemoteCommand(linkFileCmd)
}
