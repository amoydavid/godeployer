package tasks

import (
	"context"
	"testing"

	"deployer/internal/config"
	deployctx "deployer/internal/context"
)

func TestNewRollbackTask(t *testing.T) {
	tests := []struct {
		name     string
		steps    int
		expected int
	}{
		{"正常步数", 1, 1},
		{"多步回滚", 5, 5},
		{"零步（应该使用默认值1）", 0, 1},
		{"负数步数（应该使用默认值1）", -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewRollbackTask(tt.steps)
			if rollbackTask, ok := task.(*RollbackTask); ok {
				if rollbackTask.Steps != tt.expected {
					t.Errorf("expected steps %d, got %d", tt.expected, rollbackTask.Steps)
				}
			} else {
				t.Error("task should be *RollbackTask")
			}
		})
	}
}

func TestRollbackTask_Name(t *testing.T) {
	task := NewRollbackTask(1)
	if task.Name() != "rollback" {
		t.Errorf("expected name 'rollback', got '%s'", task.Name())
	}
}

func TestRollbackTask_Description(t *testing.T) {
	tests := []struct {
		steps        int
		expectedDesc string
	}{
		{1, "回滚到之前的1个发布版本"},
		{5, "回滚到之前的5个发布版本"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			task := NewRollbackTask(tt.steps)
			desc := task.Description()
			if desc != tt.expectedDesc {
				t.Errorf("expected description '%s', got '%s'", tt.expectedDesc, desc)
			}
		})
	}
}

func TestRollbackTask_SetSteps(t *testing.T) {
	task := NewRollbackTask(1)
	rollbackTask, ok := task.(*RollbackTask)
	if !ok {
		t.Fatal("task should be *RollbackTask")
	}

	// 测试设置正常值
	rollbackTask.SetSteps(5)
	if rollbackTask.Steps != 5 {
		t.Errorf("expected steps 5, got %d", rollbackTask.Steps)
	}

	// 测试设置零（应该使用默认值1）
	rollbackTask.SetSteps(0)
	if rollbackTask.Steps != 1 {
		t.Errorf("expected steps 1, got %d", rollbackTask.Steps)
	}

	// 测试设置负数（应该使用默认值1）
	rollbackTask.SetSteps(-10)
	if rollbackTask.Steps != 1 {
		t.Errorf("expected steps 1, got %d", rollbackTask.Steps)
	}
}

// 注意：完整的 Execute 方法测试需要 mock executor，
// 这里只进行基本的单元测试
func TestRollbackTask_Execute_Validation(t *testing.T) {
	// 测试边界条件
	task := NewRollbackTask(100)
	rollbackTask, ok := task.(*RollbackTask)
	if !ok {
		t.Fatal("task should be *RollbackTask")
	}

	cfg := &config.Config{
		Project: "test-project",
		Stages: map[string]config.StageConfig{
			"test": {
				Server:    "localhost",
				RemoteDir: "/tmp/test",
			},
		},
	}

	ctx, err := deployctx.NewDeployContext(context.Background(), cfg, "test", true)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	// 由于没有实际的远程服务器，这里只测试任务创建
	// 实际的 Execute 测试需要在集成测试中进行
	if rollbackTask.Steps != 100 {
		t.Errorf("expected steps 100, got %d", rollbackTask.Steps)
	}

	_ = ctx // 避免未使用变量警告
}

func TestRollbackTask_BoundaryConditions(t *testing.T) {
	// 测试各种边界条件
	tests := []struct {
		name       string
		steps      int
		validInput bool
	}{
		{"正常回滚1步", 1, true},
		{"正常回滚10步", 10, true},
		{"边界值0", 0, true},   // 会被纠正为1
		{"边界值-1", -1, true}, // 会被纠正为1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewRollbackTask(tt.steps)
			rollbackTask, ok := task.(*RollbackTask)
			if !ok {
				t.Fatal("task should be *RollbackTask")
			}

			// 验证任务被正确创建
			if task == nil {
				t.Error("task should not be nil")
			}

			// 验证步数是有效的
			if rollbackTask.Steps <= 0 {
				t.Error("steps should be positive")
			}
		})
	}
}
