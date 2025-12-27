package context

import (
	"context"
	"testing"

	"deployer/internal/config"
)

func TestResolveVar_Simple(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"name":  "test",
			"count": 42,
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 测试简单的字符串替换
	// shellescape.Quote 对安全字符串（如 "test"）不添加引号
	result := ctx.ResolveVar("Hello {{name}}")
	expected := "Hello test"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestResolveVar_MaliciousInput(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"malicious": "; rm -rf /",
			"normal":    "hello",
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 测试恶意字符被正确转义
	// shellescape.Quote 会添加单引号并转义内部的单引号
	result := ctx.ResolveVar("echo {{malicious}}")

	// 应该被转义，包含引号
	if result == "echo ; rm -rf /" {
		t.Errorf("恶意字符未被转义: %s", result)
	}

	// 验证结果包含转义后的引号
	// shellescape.Quote("; rm -rf /") 会返回 "'; rm -rf /'"
	// 或者如果包含单引号会转义为 '\'\''
	if !contains(result, "'") {
		t.Errorf("转义结果应该包含引号: %s", result)
	}
}

func TestResolveVar_SpecialCharacters(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"pipe":      "| cat",
			"and":       "&& echo hacked",
			"semicolon": "; ls",
			"backtick":  "`whoami`",
			"dollar":    "$HOME",
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 所有特殊字符都应该被转义
	testCases := []struct {
		varName string
	}{
		{"pipe"},
		{"and"},
		{"semicolon"},
		{"backtick"},
		{"dollar"},
	}

	for _, tc := range testCases {
		result := ctx.ResolveVar("test {{" + tc.varName + "}}")

		// 转义后的结果应该包含引号
		if !contains(result, "'") {
			t.Errorf("变量 %s 的值未被正确转义: %s", tc.varName, result)
		}
	}
}

func TestResolveVar_MultipleVariables(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"env":     "prod",
			"version": "1.0.0",
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	result := ctx.ResolveVar("Deploying {{version}} to {{env}}")
	expected := "Deploying 1.0.0 to prod"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestResolveVar_UnknownVariable(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"known": "value",
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 未知变量应该保持不变，已知变量应该被转义
	result := ctx.ResolveVar("{{unknown}} and {{known}}")
	expected := "{{unknown}} and value"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestResolveVar_BuiltinVariables(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Stages: map[string]config.StageConfig{
			"production": {
				Server:    "example.com",
				RemoteDir: "/var/www/app",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "production", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 测试内置变量
	tests := []struct {
		varName     string
		shouldExist bool
	}{
		{"project", true},
		{"stage", true},
		{"remote_dir", true},
		{"release_path", true},
		{"timestamp", true},
		{"keep_releases", true},
	}

	for _, tt := range tests {
		_, exists := ctx.Vars[tt.varName]
		if exists != tt.shouldExist {
			t.Errorf("变量 %s 存在性: expected %v, got %v", tt.varName, tt.shouldExist, exists)
		}
	}
}

func TestResolveVarsInMap(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"var1": "value1",
			"var2": "value2",
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	input := map[string]interface{}{
		"key1": "value {{var1}}",
		"key2": "value {{var2}}",
		"key3": 123,
		"key4": true,
	}

	result := ctx.ResolveVarsInMap(input)

	// 验证字符串值被替换
	if result["key1"].(string) != "value value1" {
		t.Errorf("expected key1 to be 'value value1', got %v", result["key1"])
	}

	if result["key2"].(string) != "value value2" {
		t.Errorf("expected key2 to be 'value value2', got %v", result["key2"])
	}

	// 验证非字符串值保持不变
	if result["key3"].(int) != 123 {
		t.Errorf("expected key3 to be 123, got %v", result["key3"])
	}

	if result["key4"].(bool) != true {
		t.Errorf("expected key4 to be true, got %v", result["key4"])
	}
}

func TestAddVar(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Vars: map[string]interface{}{
			"existing": "value",
		},
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 添加新变量
	ctx.AddVar("new_var", "new_value")

	// 验证变量被添加
	if val, ok := ctx.Vars["new_var"]; !ok || val != "new_value" {
		t.Errorf("新变量未被正确添加: %v", val)
	}

	// 覆盖现有变量
	ctx.AddVar("existing", "new_value")

	// 验证变量被覆盖
	if val, ok := ctx.Vars["existing"]; !ok || val != "new_value" {
		t.Errorf("现有变量未被正确覆盖: %v", val)
	}
}

func TestNewDeployContext_BuiltinVars(t *testing.T) {
	cfg := &config.Config{
		Project: "my-project",
		Stages: map[string]config.StageConfig{
			"prod": {
				Server:       "prod.example.com",
				RemoteDir:    "/var/www/my-app",
				KeepReleases: 5,
			},
		},
	}

	ctx, err := NewDeployContext(context.Background(), cfg, "prod", false)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 验证内置变量
	tests := []struct {
		varName  string
		expected interface{}
	}{
		{"project", "my-project"},
		{"stage", "prod"},
		{"remote_dir", "/var/www/my-app"},
		{"keep_releases", 5},
	}

	for _, tt := range tests {
		if ctx.Vars[tt.varName] != tt.expected {
			t.Errorf("内置变量 %s: expected %v, got %v",
				tt.varName, tt.expected, ctx.Vars[tt.varName])
		}
	}

	// 验证 release_path 格式
	releasePath, ok := ctx.Vars["release_path"].(string)
	if !ok {
		t.Fatal("release_path 应该是字符串")
	}

	// 应该包含时间戳
	if len(releasePath) < 20 {
		t.Errorf("release_path 格式不正确: %s", releasePath)
	}

	// 验证 timestamp 格式
	timestamp, ok := ctx.Vars["timestamp"].(string)
	if !ok || len(timestamp) != 14 {
		t.Errorf("timestamp 格式不正确: %s", timestamp)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
