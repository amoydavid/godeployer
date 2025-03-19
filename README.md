# GoDeployer

GoDeployer 是一个灵活、可扩展的部署工具，基于任务和钩子系统设计，允许用户完全定制部署流程。受 PHP Deployer 启发，但使用 Go 语言实现，提供更好的性能和跨平台支持。

## 特点

- **基于任务的部署流程**：将复杂的部署流程分解为独立、可重用的任务
- **钩子系统**：在任何任务前后执行自定义操作
- **多环境支持**：轻松在开发、测试和生产环境之间切换
- **共享资源**：支持在部署之间共享文件和目录
- **回滚支持**：保留部署历史，支持快速回滚
- **可扩展**：通过插件系统添加新功能
- **模板变量**：在配置中使用 `{{变量}}` 语法引用上下文变量
- **模拟运行**：支持 `dry-run` 模式，在不进行实际更改的情况下测试部署流程

## 安装

### 从源代码构建

```bash
git clone https://github.com/amoydavid/godeployer.git
cd godeployer
go build -o deployer ./cmd/deployer
```

然后将 `deployer` 二进制文件移动到您的 PATH 中的位置：

```bash
sudo mv deployer /usr/local/bin/
```

## 快速开始

1. 创建配置文件 `deploy.yaml`：

```yaml
project: myapp
default_stage: production

stages:
  production:
    server: user@example.com
    remote_dir: /var/www/myapp
    keep_releases: 5

options:
  shared_dirs:
    - uploads
    - logs
  shared_files:
    - .env

tasks:
  build:
    local: npm run build

recipe:
  - build
  - setup
  - update_code
  - publish_release
  - install_dependencies
  - symlink_release
  - restart_app
  - cleanup
```

2. 部署您的应用：

```bash
deployer --stage=production
```

## 配置文件

GoDeployer 使用 YAML 配置文件。详细配置选项如下：

### 基本配置

```yaml
# 项目名称
project: myapp

# 默认部署环境
default_stage: production

# 环境配置
stages:
  production:
    server: user@example.com
    remote_dir: /var/www/myapp
    keep_releases: 5
    vars:
      custom_var: "value"
  
  staging:
    server: user@staging.example.com
    remote_dir: /var/www/myapp-staging
    keep_releases: 3
```

### 全局选项

```yaml
options:
  # 发布目录格式（支持日期格式化）
  release_dir_format: "releases/%Y%m%d%H%M%S"
  
  # 共享目录（在部署之间保留）
  shared_dirs:
    - uploads
    - logs
  
  # 共享文件（在部署之间保留）
  shared_files:
    - .env
```

### 自定义任务

```yaml
tasks:
  # 构建任务
  build:
    local: npm run build
  
  # 数据库迁移任务
  migrate:
    remote: cd {{release_path}} && php artisan migrate --force
  
  # 复杂上传任务
  update_code:
    local: |
      rm -rf .deploy
      mkdir -p .deploy
      cp -r ./dist/* .deploy/
    upload:
      source: .deploy
      dest: "{{release_path}}"
      options: "--no-xattr"
```

### 钩子

```yaml
hooks:
  # 全局钩子
  before_all:
    - echo "开始部署到 {{stage}} 环境..."
  
  after_all:
    - echo "部署完成！"
  
  # 任务特定钩子
  before_restart_app:
    - echo "即将重启应用..."
  
  # 任务成功钩子
  after_publish_release:success:
    - echo "新版本已成功发布在 {{release_path}}"
    - curl -s -X POST "https://api.example.com/notify?message=发布成功"
  
  # 任务失败钩子
  after_publish_release:failed:
    - echo "发布失败！"
    - curl -s -X POST "https://api.example.com/notify?message=发布失败"
  
  # 通用任务钩子（无论成功失败）
  after_publish_release:
    - echo "发布过程已完成"
  
  # 全局失败钩子
  on_failed:
    - echo "部署失败！正在回滚..."
    - echo "失败任务: {{failed_task}}"
    - echo "错误信息: {{error}}"
```

钩子系统支持以下类型：

1. `before_all` - 所有任务执行前
2. `after_all` - 所有任务执行后
3. `before_任务名` - 特定任务执行前
4. `after_任务名` - 特定任务执行后（无论成功或失败）
5. `after_任务名:success` - 特定任务成功执行后
6. `after_任务名:failed` - 特定任务执行失败后
7. `on_failed` - 任何任务失败时

在钩子中，您可以使用所有标准变量，以及特定的任务状态信息：
- `{{task}}` - 当前任务名称
- `{{status}}` - 任务状态（'success'或'failed'）
- `{{failed_task}}` - 失败的任务名称（仅在失败钩子中）
- `{{error}}` - 错误消息（仅在失败钩子中）

### 部署流程（配方）

```yaml
# 定义部署任务执行顺序
recipe:
  - build
  - setup
  - update_code
  - publish_release
  - install_dependencies
  - migrate    # 自定义任务
  - symlink_release
  - restart_app
  - notify     # 自定义任务
  - cleanup
```

## 可用任务

GoDeployer 内置以下任务：

- **build**：在本地构建应用
- **setup**：在远程服务器上设置目录结构
- **update_code**：准备和上传代码到远程服务器
- **install_dependencies**：在远程服务器上安装应用依赖
- **publish_release**：设置共享目录和文件
- **symlink_release**: 将当前发布版本软链到current目录
- **restart_app**：重启远程服务器上的应用
- **cleanup**：清理旧的发布版本

## 变量

在配置中可以使用以下变量：

- `{{project}}`：项目名称
- `{{stage}}`：当前部署环境
- `{{remote_dir}}`：远程服务器上的项目根目录
- `{{release_path}}`：当前发布版本的完整路径
- `{{timestamp}}`：部署时间戳
- `{{keep_releases}}`：要保留的发布版本数量

## 命令行选项

```
Usage:
  -config string
    	配置文件路径 (default "deploy.yaml")
  -dry-run
    	模拟部署，不进行实际更改
  -help
    	显示帮助信息
  -plugins string
    	插件目录路径 (default "./plugins")
  -stage string
    	部署阶段（环境）
  -verbose
    	启用详细日志记录
```

### 子命令

```
deployer version      # 显示版本信息
deployer init         # 初始化配置文件
deployer list         # 列出配置的任务和环境
deployer rollback     # 回滚到之前的版本
  -steps int          # 回滚的步数，默认为1 (回滚到上一个版本)
```

## 插件开发

GoDeployer 支持通过插件扩展功能。插件是实现 `Plugin` 接口的 Go 共享库（.so 文件）。

插件接口定义：

```go
type Plugin interface {
	// Name 返回插件名称
	Name() string
	
	// Version 返回插件版本
	Version() string
	
	// Register 向任务注册表注册插件任务
	Register(registry *registry.TaskRegistry) error
}
```

## 许可证

GNU Affero General Public License v3.0 (AGPL-3.0)

该项目根据GNU Affero通用公共许可证第3版授权。这意味着如果您修改该软件并通过网络提供服务，您必须提供修改后的源代码。完整的许可证文本可在LICENSE文件中找到。

## 贡献

欢迎提交问题和拉取请求！ 