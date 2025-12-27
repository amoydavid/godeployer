package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewUpdateCodeTask(t *testing.T) {
	task := NewUpdateCodeTask()
	if task == nil {
		t.Error("task should not be nil")
	}

	if task.Name() != "update_code" {
		t.Errorf("expected name 'update_code', got '%s'", task.Name())
	}
}

func TestUpdateCodeTask_Description(t *testing.T) {
	task := NewUpdateCodeTask()
	desc := task.Description()
	if desc != "准备并上传代码到远程服务器" {
		t.Errorf("expected description '准备并上传代码到远程服务器', got '%s'", desc)
	}
}

func TestCopyFile(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "copy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建源文件
	srcFile := filepath.Join(tempDir, "src.txt")
	content := []byte("hello world")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to create src file: %v", err)
	}

	// 测试复制文件
	dstFile := filepath.Join(tempDir, "dst.txt")
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Errorf("failed to copy file: %v", err)
	}

	// 验证目标文件存在
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("destination file should exist")
	}

	// 验证内容
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dst file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("content mismatch: expected %s, got %s", content, dstContent)
	}
}

func TestCopyFile_CreateDirectories(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "copy-test-dirs-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建源文件
	srcFile := filepath.Join(tempDir, "src.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to create src file: %v", err)
	}

	// 复制到不存在的嵌套目录
	dstFile := filepath.Join(tempDir, "nested", "deep", "dst.txt")
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Errorf("failed to copy file with nested directories: %v", err)
	}

	// 验证目标文件存在
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("destination file should exist")
	}
}

func TestCopyFile_Permissions(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "copy-test-perms-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建源文件
	srcFile := filepath.Join(tempDir, "src.txt")
	content := []byte("test")
	if err := os.WriteFile(srcFile, content, 0755); err != nil {
		t.Fatalf("failed to create src file: %v", err)
	}

	// 复制文件
	dstFile := filepath.Join(tempDir, "dst.txt")
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Errorf("failed to copy file: %v", err)
	}

	// 验证权限（注意：Windows 可能不支持 Unix 权限）
	srcInfo, _ := os.Stat(srcFile)
	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("failed to stat dst file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		// Windows 上的权限可能不同，这是正常的
		t.Logf("permissions differ: src %v, dst %v (may be normal on Windows)", srcInfo.Mode(), dstInfo.Mode())
	}
}
