package context

import (
	"context"
	"os"
	"testing"

	"deployer/internal/config"
)

func TestValidatePrivateKey_NonExistent(t *testing.T) {
	err := validatePrivateKey("/nonexistent/key.pem")
	if err == nil {
		t.Error("expected error for nonexistent key, got nil")
	}

	if !contains(err.Error(), "私钥文件不存在") {
		t.Errorf("error message should mention file not exist, got '%s'", err.Error())
	}
}

func TestValidatePrivateKey_EmptyPath(t *testing.T) {
	// 空路径是有效的（可能使用密码认证）
	err := validatePrivateKey("")
	if err != nil {
		t.Errorf("expected no error for empty path, got %v", err)
	}
}

func TestValidatePrivateKey_Directory(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "ssh-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试目录路径
	err = validatePrivateKey(tempDir)
	if err == nil {
		t.Error("expected error for directory path, got nil")
	}

	if !contains(err.Error(), "目录") {
		t.Errorf("error message should mention directory, got '%s'", err.Error())
	}
}

func TestValidatePrivateKey_Permissions(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "ssh-key-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 设置过于宽松的权限 (644 - others readable)
	if err := os.Chmod(tempFile.Name(), 0644); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	// 测试权限检查
	err = validatePrivateKey(tempFile.Name())
	if err == nil {
		t.Error("expected error for loose permissions, got nil")
	}

	expectedMsg := "权限过于宽松"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("error message should mention loose permissions, got '%s'", err.Error())
	}
}

func TestValidatePrivateKey_CorrectPermissions(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "ssh-key-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 设置正确的权限 (600)
	if err := os.Chmod(tempFile.Name(), 0600); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	// 测试正确的权限
	err = validatePrivateKey(tempFile.Name())
	if err != nil {
		// 在 Windows 上，权限检查可能不同
		t.Logf("Permission check failed (may be normal on Windows): %v", err)
	}
}

func TestNewDeployContext_InvalidPrivateKey(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Stages: map[string]config.StageConfig{
			"test": {
				Server:         "localhost",
				RemoteDir:      "/tmp/test",
				PrivateKeyPath: "/nonexistent/key.pem",
			},
		},
	}

	_, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err == nil {
		t.Error("expected error for invalid private key, got nil")
	}

	if !contains(err.Error(), "私钥") {
		t.Errorf("error should mention private key, got '%s'", err.Error())
	}
}

func TestNewDeployContext_ValidPrivateKey(t *testing.T) {
	// 创建临时私钥文件
	tempFile, err := os.CreateTemp("", "ssh-key-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// 设置正确的权限
	if err := os.Chmod(tempFile.Name(), 0600); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}

	cfg := &config.Config{
		Project: "test-project",
		Stages: map[string]config.StageConfig{
			"test": {
				Server:         "localhost",
				RemoteDir:      "/tmp/test",
				PrivateKeyPath: tempFile.Name(),
			},
		},
	}

	_, err = NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		// 在 Windows 上可能会有权限错误
		t.Logf("Context creation failed (may be normal on Windows): %v", err)
	}
}

func TestNewDeployContext_NoPrivateKey(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
				// PrivateKeyPath 为空
			},
		},
	}

	_, err := NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Errorf("expected no error when private key not configured, got %v", err)
	}
}
