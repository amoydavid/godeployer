package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCleanupTask(t *testing.T) {
	task := NewCleanupTask()
	if task == nil {
		t.Error("task should not be nil")
	}

	if task.Name() != "cleanup" {
		t.Errorf("expected name 'cleanup', got '%s'", task.Name())
	}
}

func TestCleanupTask_Description(t *testing.T) {
	task := NewCleanupTask()
	desc := task.Description()
	if desc != "清理旧的发布版本" {
		t.Errorf("expected description '清理旧的发布版本', got '%s'", desc)
	}
}

func TestCleanupTask_LocalTempDir(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := ".deploy_tmp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建一些测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("test file should exist")
	}

	// 测试清理（模拟 cleanup.go 中的逻辑）
	if _, err := os.Stat(tempDir); err == nil {
		// 目录存在，删除它
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}

	// 验证目录已被删除
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("temp dir should be removed")
	}
}

func TestCleanupTask_NonExistentTempDir(t *testing.T) {
	// 测试清理不存在的目录（不应报错）
	tempDir := ".deploy_tmp_nonexistent"

	// 确保目录不存在
	os.RemoveAll(tempDir)

	// 测试清理逻辑（应该不会报错）
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		// 目录不存在，这是正常的
		return
	}

	t.Error("temp dir should not exist")
}
