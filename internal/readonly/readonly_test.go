package readonly

import (
	"testing"
)

func TestIsReadOnly_False(t *testing.T) {
	cfg := &mockConfig{readOnly: false}
	if IsReadOnly(cfg) {
		t.Error("expected false when config.ReadOnly is false")
	}
}

func TestIsReadOnly_True(t *testing.T) {
	cfg := &mockConfig{readOnly: true}
	if !IsReadOnly(cfg) {
		t.Error("expected true when config.ReadOnly is true")
	}
}

func TestIsReadOnly_NilConfig(t *testing.T) {
	if IsReadOnly(nil) {
		t.Error("expected false when config is nil")
	}
}

type mockConfig struct {
	readOnly bool
}

func (m *mockConfig) ReadOnly() bool {
	return m.readOnly
}
