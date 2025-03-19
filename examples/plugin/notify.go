package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	deployctx "deployer/internal/context"
	"deployer/internal/registry"
)

// NotifyPlugin 是一个通知插件
type NotifyPlugin struct{}

// NotifyTask 是一个发送通知的任务
type NotifyTask struct {
	webhookURL string
}

// Name 返回任务名称
func (t *NotifyTask) Name() string {
	return "notify"
}

// Description 返回任务描述
func (t *NotifyTask) Description() string {
	return "向Slack或其他Webhook发送部署通知"
}

// Execute 执行通知任务
func (t *NotifyTask) Execute(ctx *deployctx.DeployContext) error {
	// 从配置中获取Webhook URL
	taskCfg, ok := ctx.Config.Tasks["notify"]
	if ok && taskCfg.Options != nil {
		if url, exists := taskCfg.Options["webhook_url"]; exists {
			t.webhookURL = ctx.ResolveVar(url)
		}
	}

	if t.webhookURL == "" {
		ctx.Logger.Warn("没有配置Webhook URL，跳过通知")
		return nil
	}

	// 构建通知消息
	message := fmt.Sprintf("项目 %s 已成功部署到 %s 环境！\n发布路径: %s",
		ctx.Config.Project, ctx.Stage, ctx.ReleasePath)

	// 发送通知
	data := url.Values{}
	data.Set("payload", fmt.Sprintf(`{"text": "%s"}`, message))

	resp, err := http.Post(t.webhookURL, "application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))

	if err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("发送通知失败，状态码: %d", resp.StatusCode)
	}

	ctx.Logger.Success("通知已发送")
	return nil
}

// Name 返回插件名称
func (p *NotifyPlugin) Name() string {
	return "notify"
}

// Version 返回插件版本
func (p *NotifyPlugin) Version() string {
	return "1.0.0"
}

// Register 注册插件任务
func (p *NotifyPlugin) Register(registry *registry.TaskRegistry) error {
	return registry.Register(&NotifyTask{})
}

// Plugin 是插件的导出符号
var Plugin = &NotifyPlugin{}
