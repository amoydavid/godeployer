package executor

import (
	"testing"
)

func TestSanitizeServerName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "user_example.com"},       // @ → _
		{"192.168.1.1", "192.168.1.1"},               // IP地址保持不变
		{"server-name.domain.com", "server-name.domain.com"}, // - 和 . 保留
		{"user@host:22", "user_host_22"},               // @ 和 : → _
		{"simple-host", "simple-host"},                // 连字符保留
		{"server_with_underscore", "server_with_underscore"}, // 下划线保留
		{"test@server.com:2222", "test_server.com_2222"}, // 多个特殊字符，点号保留
	}

	for _, tt := range tests {
		result := sanitizeServerName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeServerName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewSSHPool(t *testing.T) {
	pool := NewSSHPool()
	if pool == nil {
		t.Fatal("NewSSHPool() returned nil")
	}

	if pool.sockets == nil {
		t.Error("pool.sockets is nil")
	}
}

func TestGetGlobalPool(t *testing.T) {
	pool1 := GetGlobalPool()
	pool2 := GetGlobalPool()

	if pool1 != pool2 {
		t.Error("GetGlobalPool() should return the same instance")
	}
}
