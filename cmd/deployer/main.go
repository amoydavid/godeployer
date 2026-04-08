/*
 * GoDeployer - 灵活、可扩展的部署工具
 * Copyright (C) 2025-2025 amoydavid
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"context"
	"deployer/internal/config"
	deployctx "deployer/internal/context"
	"deployer/internal/example"
	"deployer/internal/executor"
	"deployer/internal/recipe"
	"deployer/internal/registry"
	"deployer/internal/tasks"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/spf13/cobra"
)

// 版本信息
const (
	VERSION = "0.1.0"
)

var (
	configFile    string
	stage         string
	dryRun        bool
	pluginDir     string
	rollbackSteps int
	RootCmd       = &cobra.Command{
		Use:   "deployer",
		Short: "GoDeployer - 灵活的部署工具",
		Long:  `GoDeployer - 一个灵活、可扩展的部署自动化工具。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy()
		},
	}
)

func main() {
	// 添加命令行标志
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "deploy.yaml", "配置文件路径")
	RootCmd.PersistentFlags().StringVar(&stage, "stage", "", "部署阶段（环境）")
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "模拟部署，不进行实际更改")
	// RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "启用详细日志记录")
	RootCmd.PersistentFlags().StringVar(&pluginDir, "plugins", "./plugins", "插件目录路径 (当前版本中已禁用)")

	// 添加版本子命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GoDeployer 版本 %s\n", VERSION)
		},
	}
	RootCmd.AddCommand(versionCmd)

	// 添加初始化配置子命令
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "初始化一个部署配置文件",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}
	RootCmd.AddCommand(initCmd)

	// 添加任务列表子命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出配置文件中定义的任务和部署阶段",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTasks()
		},
	}
	RootCmd.AddCommand(listCmd)

	// 添加回滚子命令
	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "回滚到之前的发布版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRollback()
		},
	}
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "回滚的步数，默认为1，表示回滚到上一个版本")
	RootCmd.AddCommand(rollbackCmd)

	// 添加 SSH 子命令
	sshCmd := &cobra.Command{
		Use:   "ssh [stage]",
		Short: "通过 SSH 登录到远程服务器",
		Long: `通过 SSH 登录到远程服务器，获得完整的 shell 环境。
如果没有指定 stage，将使用默认 stage。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSSH(cmd, args)
		},
	}
	RootCmd.AddCommand(sshCmd)

	// 添加 example 子命令
	exampleCmd := &cobra.Command{
		Use:   "example <filename>",
		Short: "Generate interactive example configuration file",
		Long: `Generate an example deployment configuration file through interactive menus.

This command will guide you through creating a customized deployment configuration
by asking you a series of questions about your project.

CONFIGURATION COMPLEXITY LEVELS:

  simple   - Minimal configuration for static sites or simple applications
             Includes: basic project info, dev/prod stages, simple recipe

  standard - Typical web application configuration
             Includes: tasks, build commands, shared directories/files,
                      standard deployment recipe, basic hooks

  advanced - Multi-environment configuration with complex hooks
             Includes: dev/staging/prod stages, backup tasks,
                      migration tasks, advanced hooks (success/failed),
                      shared resources, environment-specific variables

  full     - Complete demonstration of all features
             Includes: everything from advanced plus comprehensive
                      documentation, notification integrations,
                      all available configuration options

SUPPORTED PROJECT TYPES:

  nodejs  - Node.js applications (npm build)
  go      - Go applications (go build)
  python  - Python applications (python -m build)
  php     - PHP applications (composer install)
  static  - Static websites (no build needed)
  generic - Generic/custom projects

EXAMPLES:

  # Generate an interactive configuration with prompts
  deployer example deploy.yaml

  # The command will ask you to:
  # 1. Choose complexity level (simple/standard/advanced/full)
  # 2. Select project type (nodejs/go/python/php/static/generic)
  # 3. Whether to include detailed comments (yes/no)

OUTPUT FILE STRUCTURE:

  project: my-app          # Project name
  default_stage: dev        # Default deployment environment

  stages:                   # Deployment environments
    dev:
      server: server.com
      remote_dir: /var/www/app
      keep_releases: 3
      private_key: ~/.ssh/id_rsa
      vars:                  # Environment variables
        APP_ENV: development

  options:                  # Global options
    release_dir_format: "releases/%Y%m%d%H%M%S"
    shared_dirs:            # Directories shared across releases
      - logs
      - uploads
    shared_files:           # Files shared across releases
      - .env

  tasks:                    # Custom deployment tasks
    build:
      local: npm run build

  recipe:                   # Deployment workflow steps
    - build
    - update_code
    - symlink_release
    - cleanup

  hooks:                    # Lifecycle hooks
    before_all:
      - echo "Starting deployment"
    after_build:success:
      - echo "Build successful!"
    on_failed:
      - echo "Deployment failed"

  vars:                     # Global variables
    app_name: my-app
    node_version: "18"

For more information, visit: https://github.com/yourusername/godeployer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return example.Run(args)
		},
	}
	RootCmd.AddCommand(exampleCmd)

	// 执行命令
	if err := RootCmd.Execute(); err != nil {
		// Cobra 会处理错误退出
		os.Exit(1)
	}
}

// runSSH 执行 SSH 登录
func runSSH(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置错误: %w", err)
	}

	// 确定要使用的阶段
	selectedStage := ""
	if len(args) > 0 {
		selectedStage = args[0]
	} else {
		selectedStage = cfg.DefaultStage
	}

	if _, exists := cfg.Stages[selectedStage]; !exists {
		return fmt.Errorf("配置中未找到阶段 '%s'", selectedStage)
	}

	// 创建部署上下文
	deployCtx, err := deployctx.NewDeployContext(cmd.Context(), cfg, selectedStage, false)
	if err != nil {
		return fmt.Errorf("创建部署上下文错误: %w", err)
	}

	// 构建 SSH 命令，支持私钥路径
	sshArgs := []string{"-t"}
	sshArgs = append(sshArgs, executor.BuildSSHArgs(deployCtx.StageConfig.PrivateKeyPath, deployCtx.StageConfig.Port)...)
	sshArgs = append(sshArgs, deployCtx.StageConfig.Server,
		"cd "+deployCtx.StageConfig.RemoteDir+"/current 2>/dev/null || cd "+deployCtx.StageConfig.RemoteDir+"; exec $SHELL -l",
	)

	// 执行 SSH 命令
	return executor.RunSSHCommand(sshArgs...)
}

// 列出配置文件中定义的任务和部署阶段
func listTasks() error {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置错误: %w", err)
	}

	fmt.Printf("项目: %s\n", cfg.Project)
	fmt.Printf("默认部署阶段: %s\n\n", cfg.DefaultStage)

	// 列出部署阶段
	fmt.Println("可用部署阶段:")
	for stageName := range cfg.Stages {
		if stageName == cfg.DefaultStage {
			fmt.Printf("  - %s (默认)\n", stageName)
		} else {
			fmt.Printf("  - %s\n", stageName)
		}
	}
	fmt.Println()

	// 列出自定义任务
	fmt.Println("自定义任务:")
	if len(cfg.Tasks) == 0 {
		fmt.Println("  未定义自定义任务")
	} else {
		taskNames := make([]string, 0, len(cfg.Tasks))
		for name := range cfg.Tasks {
			taskNames = append(taskNames, name)
		}
		sort.Strings(taskNames)

		for _, name := range taskNames {
			taskCfg := cfg.Tasks[name]
			description := "未提供描述"

			if taskCfg.Local != "" {
				description = fmt.Sprintf("本地命令: %s", taskCfg.Local)
			} else if taskCfg.Remote != "" {
				description = fmt.Sprintf("远程命令: %s", taskCfg.Remote)
			} else if taskCfg.Upload.Source != "" {
				description = fmt.Sprintf("上传: %s -> %s", taskCfg.Upload.Source, taskCfg.Upload.Dest)
			}

			fmt.Printf("  - %s: %s\n", name, description)
		}
	}
	fmt.Println()

	// 列出部署配方
	fmt.Println("部署步骤:")
	if len(cfg.Recipe) == 0 {
		fmt.Println("  未定义部署步骤")
	} else {
		for i, task := range cfg.Recipe {
			fmt.Printf("  %d. %s\n", i+1, task)
		}
	}
	return nil
}

// 初始化一个示例配置文件
func initConfig() error {
	// 检查配置文件是否已存在
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("配置文件 '%s' 已存在，要覆盖它吗？(y/n): ", configFile)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("操作已取消")
			return nil
		}
	}

	// 创建示例配置
	exampleConfig := `# GoDeployer 配置示例
project: my-web-app
default_stage: dev

# 定义不同部署环境
stages:
  dev:
    server: dev-server.example.com
    remote_dir: /var/www/my-app-dev
    keep_releases: 3
    vars:
      APP_ENV: development
      DEBUG: "true"

  prod:
    server: prod-server.example.com
    remote_dir: /var/www/my-app-prod
    keep_releases: 5
    vars:
      APP_ENV: production
      DEBUG: "false"

# 全局选项
options:
  release_dir_format: "releases/%Y%m%d%H%M%S"
  shared_dirs:
    - logs
    - uploads
  shared_files:
    - .env

# 自定义任务
tasks:
  build:
    local: npm run build

  backup:
    remote: cp -r /var/www/app /var/www/app_backup_$(date +%Y%m%d)

  upload-app:
    upload:
      source: ./dist
      dest: "{{release_path}}"
      options: "--delete"

# 部署步骤
recipe:
  - build
  - backup
  - upload-app
  - publish_release
  - symlink_release
  - restart_app
  - cleanup

# 部署钩子
hooks:
  # 全局钩子
  before_all:
    - echo "开始部署到 {{stage}} 环境..."

  after_all:
    - echo "部署完成！"

  # 任务特定前置钩子
  before_build:
    - echo "准备开始构建..."

  # 任务特定后置钩子（无论成功失败）
  after_build:
    - echo "构建过程已完成"

  # 任务成功钩子
  after_build:success:
    - echo "构建成功完成！"
    - echo "通知团队构建成功"

  # 任务失败钩子
  after_build:failed:
    - echo "构建失败！"
    - echo "通知团队构建失败"

  after_upload-app:success:
    - echo "文件上传成功，新版本已位于 {{release_path}}"

  after_backup:failed:
    - echo "备份失败，但将继续部署过程..."

  # 全局失败钩子
  on_failed:
    - echo "部署过程中出现错误！错误信息: {{error}}"

# 全局变量
vars:
  app_name: my-awesome-app
  node_version: "16"
`

	// 写入配置文件
	if err := os.WriteFile(configFile, []byte(exampleConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("示例配置已写入 '%s'\n", configFile)
	return nil
}

func runDeploy() error {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置错误: %w", err)
	}

	// 确定要使用的阶段
	selectedStage := stage
	if selectedStage == "" {
		selectedStage = cfg.DefaultStage
	}

	if _, exists := cfg.Stages[selectedStage]; !exists {
		return fmt.Errorf("配置中未找到阶段 '%s'", selectedStage)
	}

	// 用户确认
	dryRunText := ""
	if dryRun {
		dryRunText = "（模拟运行模式）"
	}
	fmt.Printf("将部署 %s 到 %s 环境%s，确认继续? (y/n): ", cfg.Project, selectedStage, dryRunText)
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("部署已取消")
		return nil
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号进行优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n收到中断信号，正在优雅关闭...")
		cancel()
	}()

	// 创建部署上下文
	deployCtx, err := deployctx.NewDeployContext(ctx, cfg, selectedStage, dryRun)
	if err != nil {
		return fmt.Errorf("创建部署上下文错误: %w", err)
	}

	// 创建任务注册表
	taskRegistry := registry.NewTaskRegistry()

	// 注册内置任务
	tasks.RegisterBuiltinTasks(taskRegistry)

	// 注册配置中的自定义任务
	for name, taskCfg := range cfg.Tasks {
		registerCustomTaskFromConfig(taskRegistry, name, taskCfg)
	}

	// 创建并运行配方
	deployRecipe := recipe.NewRecipe(taskRegistry, cfg.Recipe, cfg.Hooks)

	fmt.Printf("正在部署 %s 到 %s 环境\n", cfg.Project, selectedStage)
	if dryRun {
		fmt.Println("模拟运行: 不会进行任何更改")
	}

	if err := deployRecipe.Execute(deployCtx); err != nil {
		return fmt.Errorf("部署失败: %w", err)
	}

	fmt.Println("部署成功完成!")
	return nil
}

func registerCustomTaskFromConfig(registry *registry.TaskRegistry, name string, taskCfg config.TaskConfig) {
	// 基于配置创建任务函数
	taskFn := func(ctx *deployctx.DeployContext) error {
		exec := executor.NewExecutor(ctx)

		// 如果定义了本地命令，则执行
		if taskCfg.Local != "" {
			if err := exec.RunLocalCommand(taskCfg.Local); err != nil {
				return fmt.Errorf("本地命令失败: %w", err)
			}
		}

		// 如果定义了远程命令，则执行
		if taskCfg.Remote != "" {
			if err := exec.RunRemoteCommand(taskCfg.Remote); err != nil {
				return fmt.Errorf("远程命令失败: %w", err)
			}
		}

		// 如果定义了上传，则执行
		if taskCfg.Upload.Source != "" && taskCfg.Upload.Dest != "" {
			if err := exec.UploadDirectory(
				taskCfg.Upload.Source,
				taskCfg.Upload.Dest,
				taskCfg.Upload.Options,
			); err != nil {
				return fmt.Errorf("上传失败: %w", err)
			}
		}

		return nil
	}

	// 注册自定义任务
	registry.RegisterCustomTask(name, fmt.Sprintf("自定义任务: %s", name), taskFn)
}

// runRollback 执行回滚操作
func runRollback() error {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置错误: %w", err)
	}

	// 确定要使用的阶段
	selectedStage := stage
	if selectedStage == "" {
		selectedStage = cfg.DefaultStage
	}

	if _, exists := cfg.Stages[selectedStage]; !exists {
		return fmt.Errorf("配置中未找到阶段 '%s'", selectedStage)
	}

	// 用户确认
	rollbackDryRunText := ""
	if dryRun {
		rollbackDryRunText = "（模拟运行模式）"
	}
	fmt.Printf("将回滚 %s 环境，步数：%d%s，确认继续? (y/n): ", selectedStage, rollbackSteps, rollbackDryRunText)
	var rollbackResponse string
	fmt.Scanln(&rollbackResponse)
	if rollbackResponse != "y" && rollbackResponse != "Y" {
		fmt.Println("回滚已取消")
		return nil
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号进行优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n收到中断信号，正在优雅关闭...")
		cancel()
	}()

	// 创建部署上下文
	deployCtx, err := deployctx.NewDeployContext(ctx, cfg, selectedStage, dryRun)
	if err != nil {
		return fmt.Errorf("创建部署上下文错误: %w", err)
	}

	// 创建任务注册表
	taskRegistry := registry.NewTaskRegistry()

	// 注册内置任务
	tasks.RegisterBuiltinTasks(taskRegistry)

	// 获取回滚任务
	rollbackTask, err := taskRegistry.Get("rollback")
	if err != nil {
		return fmt.Errorf("加载回滚任务失败: %w", err)
	}

	// 设置回滚步数
	if rt, ok := rollbackTask.(*tasks.RollbackTask); ok {
		rt.SetSteps(rollbackSteps)
	} else {
		// 不是预期的类型，打印警告
		fmt.Fprintf(os.Stderr, "警告: 无法设置回滚步数，使用默认值\n")
	}

	fmt.Printf("正在回滚 %s 环境，步数：%d\n", selectedStage, rollbackSteps)
	if dryRun {
		fmt.Println("模拟运行: 不会进行任何更改")
	}

	// 直接执行回滚任务
	if err := rollbackTask.Execute(deployCtx); err != nil {
		return fmt.Errorf("回滚失败: %w", err)
	}

	fmt.Println("回滚成功完成!")
	return nil
}
