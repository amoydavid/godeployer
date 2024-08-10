package tasks

import (
	"fmt"
	"github.com/amoydavid/godeployer/internal/config"
	"github.com/amoydavid/godeployer/internal/ssh"
)

// Task 表示一个部署任务
type Task struct {
	Name        string
	Description string
	Run         func(*ssh.Client, *config.Config) error
}

// 定义任务列表
var DeployTasks = []Task{
	{
		Name:        "Prepare",
		Description: "Prepare the deployment environment",
		Run:         PrepareTask,
	},
	{
		Name:        "Update Code",
		Description: "Pull the latest code from the repository",
		Run:         UpdateCodeTask,
	},
	{
		Name:        "Install Dependencies",
		Description: "Install or update project dependencies",
		Run:         InstallDependenciesTask,
	},
	{
		Name:        "Build",
		Description: "Build the project",
		Run:         BuildTask,
	},
	{
		Name:        "Migrate Database",
		Description: "Run database migrations",
		Run:         MigrateDatabaseTask,
	},
	{
		Name:        "Symlink",
		Description: "Update the symlink to the new release",
		Run:         SymlinkTask,
	},
	{
		Name:        "Restart Services",
		Description: "Restart application services",
		Run:         RestartServicesTask,
	},
}

// PrepareTask 准备部署环境
func PrepareTask(client *ssh.Client, cfg *config.Config) error {
	cmd := fmt.Sprintf("mkdir -p %s/releases", cfg.DeployPath)
	_, err := client.Run(cmd)
	return err
}

// UpdateCodeTask 更新代码
func UpdateCodeTask(client *ssh.Client, cfg *config.Config) error {
	cmd := fmt.Sprintf("git clone %s %s/releases/$(date +%%Y%%m%%d%%H%%M%%S)", cfg.Repository, cfg.DeployPath)
	_, err := client.Run(cmd)
	return err
}

// InstallDependenciesTask 安装依赖
func InstallDependenciesTask(client *ssh.Client, cfg *config.Config) error {
	cmd := fmt.Sprintf("cd %s/releases/$(ls -t %s/releases | head -n1) && npm install", cfg.DeployPath, cfg.DeployPath)
	_, err := client.Run(cmd)
	return err
}

// BuildTask 构建项目
func BuildTask(client *ssh.Client, cfg *config.Config) error {
	cmd := fmt.Sprintf("cd %s/releases/$(ls -t %s/releases | head -n1) && npm run build", cfg.DeployPath, cfg.DeployPath)
	_, err := client.Run(cmd)
	return err
}

// MigrateDatabaseTask 运行数据库迁移
func MigrateDatabaseTask(client *ssh.Client, cfg *config.Config) error {
	cmd := fmt.Sprintf("cd %s/releases/$(ls -t %s/releases | head -n1) && npm run migrate", cfg.DeployPath, cfg.DeployPath)
	_, err := client.Run(cmd)
	return err
}

// SymlinkTask 更新符号链接
func SymlinkTask(client *ssh.Client, cfg *config.Config) error {
	cmd := fmt.Sprintf("ln -snf %s/releases/$(ls -t %s/releases | head -n1) %s/current", cfg.DeployPath, cfg.DeployPath, cfg.DeployPath)
	_, err := client.Run(cmd)
	return err
}

// RestartServicesTask 重启服务
func RestartServicesTask(client *ssh.Client, cfg *config.Config) error {
	cmd := "sudo systemctl restart nginx"
	_, err := client.Run(cmd)
	return err
}

// Deploy 执行完整的部署流程
func Deploy(client *ssh.Client, cfg *config.Config) error {
	for _, task := range DeployTasks {
		fmt.Printf("Executing task: %s\n", task.Name)
		if err := task.Run(client, cfg); err != nil {
			return fmt.Errorf("Task '%s' failed: %w", task.Name, err)
		}
	}
	return nil
}
