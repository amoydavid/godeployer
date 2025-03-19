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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"deployer/internal/config"
	deployctx "deployer/internal/context"
	"deployer/internal/executor"
	"deployer/internal/recipe"
	"deployer/internal/registry"
	"deployer/internal/tasks"

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
)

func main() {
	// 创建根命令
	rootCmd := &cobra.Command{
		Use:   "deployer",
		Short: "GoDeployer - 灵活的部署工具",
		Long:  `GoDeployer - 一个灵活、可扩展的部署自动化工具。`,
		Run: func(cmd *cobra.Command, args []string) {
			runDeploy()
		},
	}

	// 添加命令行标志
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "deploy.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVar(&stage, "stage", "", "部署阶段（环境）")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "模拟部署，不进行实际更改")
	// rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "启用详细日志记录")
	rootCmd.PersistentFlags().StringVar(&pluginDir, "plugins", "./plugins", "插件目录路径 (当前版本中已禁用)")

	// 添加版本子命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GoDeployer 版本 %s\n", VERSION)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// 添加初始化配置子命令
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "初始化一个部署配置文件",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
		},
	}
	rootCmd.AddCommand(initCmd)

	// 添加任务列表子命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出配置文件中定义的任务和部署阶段",
		Run: func(cmd *cobra.Command, args []string) {
			listTasks()
		},
	}
	rootCmd.AddCommand(listCmd)

	// 添加回滚子命令
	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "回滚到之前的发布版本",
		Run: func(cmd *cobra.Command, args []string) {
			runRollback()
		},
	}
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "回滚的步数，默认为1，表示回滚到上一个版本")
	rootCmd.AddCommand(rollbackCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "执行出错: %v\n", err)
		os.Exit(1)
	}
}

// 列出配置文件中定义的任务和部署阶段
func listTasks() {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置错误: %v\n", err)
		os.Exit(1)
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
}

// 初始化一个示例配置文件
func initConfig() {
	// 检查配置文件是否已存在
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("配置文件 '%s' 已存在，要覆盖它吗？(y/n): ", configFile)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("操作已取消")
			return
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
	err := os.WriteFile(configFile, []byte(exampleConfig), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "写入配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("示例配置已写入 '%s'\n", configFile)
}

func runDeploy() {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置错误: %v\n", err)
		os.Exit(1)
	}

	// 确定要使用的阶段
	selectedStage := stage
	if selectedStage == "" {
		selectedStage = cfg.DefaultStage
	}

	if _, exists := cfg.Stages[selectedStage]; !exists {
		fmt.Fprintf(os.Stderr, "错误: 配置中未找到阶段 '%s'\n", selectedStage)
		os.Exit(1)
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
		return
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
		fmt.Fprintf(os.Stderr, "创建部署上下文错误: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "部署失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("部署成功完成!")
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
func runRollback() {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置错误: %v\n", err)
		os.Exit(1)
	}

	// 确定要使用的阶段
	selectedStage := stage
	if selectedStage == "" {
		selectedStage = cfg.DefaultStage
	}

	if _, exists := cfg.Stages[selectedStage]; !exists {
		fmt.Fprintf(os.Stderr, "错误: 配置中未找到阶段 '%s'\n", selectedStage)
		os.Exit(1)
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
		return
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
		fmt.Fprintf(os.Stderr, "创建部署上下文错误: %v\n", err)
		os.Exit(1)
	}

	// 创建任务注册表
	taskRegistry := registry.NewTaskRegistry()

	// 注册内置任务
	tasks.RegisterBuiltinTasks(taskRegistry)

	// 获取回滚任务
	rollbackTask, err := taskRegistry.Get("rollback")
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载回滚任务失败: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "回滚失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("回滚成功完成!")
}
