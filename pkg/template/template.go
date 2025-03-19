package template

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderString 使用给定的上下文变量渲染模板字符串
func RenderString(templateStr string, vars map[string]interface{}) (string, error) {
	// 创建新模板
	tmpl, err := template.New("inline").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return buf.String(), nil
}

// RenderFile 读取模板文件并使用给定的上下文变量渲染它
func RenderFile(templateFile string, vars map[string]interface{}) (string, error) {
	// 解析模板文件
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", fmt.Errorf("解析模板文件 %s 失败: %w", templateFile, err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return buf.String(), nil
}
