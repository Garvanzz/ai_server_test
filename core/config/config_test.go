// Package config 的测试与测试工具。
//
// 单元测试使用 testdata/ 下的最小 JSON（0Global.json、17MonthCard.json），
// 不依赖项目真实配置目录即可运行。
//
// 运行方式:
//
//	go test ./core/config -v
//	go test ./core/config -run TestLoadRealConfigDir -v   # 用真实 json 目录做集成校验
//	CONFIG_DIR=/path/to/json go test ./core/config -run TestLoadRealConfigDir -v
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// getTestdataDir 返回 testdata 目录的绝对路径（与当前测试文件同目录下的 testdata）。
func getTestdataDir(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "testdata")
}

func TestNewManager(t *testing.T) {
	dir := getTestdataDir(t)
	m := NewManager(dir, false, Parse)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.filePath != dir {
		t.Errorf("filePath = %q, want %q", m.filePath, dir)
	}
	if m.parserFunc == nil {
		t.Error("parserFunc is nil")
	}
	all := m.getAll()
	if all == nil || len(all) != 0 {
		t.Errorf("initial getAll() should be empty map, got len=%d", len(all))
	}
}

func TestManager_LoadFile(t *testing.T) {
	dir := getTestdataDir(t)
	m := NewManager(dir, false, Parse)
	data := m.loadFile()
	if len(data) == 0 {
		t.Fatal("loadFile returned empty map, testdata should contain 0Global.json")
	}
	if _, ok := data["Global"]; !ok {
		t.Errorf("loadFile should contain key Global, got keys: %v", keys(data))
	}
}

func TestParse_Success(t *testing.T) {
	dir := getTestdataDir(t)
	m := NewManager(dir, false, Parse)
	raw := m.loadFile()
	dest := make(map[string]any)
	err := Parse(raw, dest)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if _, ok := dest["Global"]; !ok {
		t.Errorf("dest should contain Global after Parse, got keys: %v", keys(dest))
	}
}

func TestParse_MissingRegistered(t *testing.T) {
	// 仅提供空 data，缺少 registry 中注册的 Global，应返回 error
	data := make(map[string]any)
	dest := make(map[string]any)
	err := Parse(data, dest)
	if err == nil {
		t.Fatal("Parse with missing registered config should return error")
	}
}

func TestManager_Reload(t *testing.T) {
	dir := getTestdataDir(t)
	m := NewManager(dir, false, Parse)
	m.Reload()
	all := m.getAll()
	if _, ok := all["Global"]; !ok {
		t.Fatalf("after Reload, getAll() should contain Global, got keys: %v", keys(all))
	}
}

func TestManager_Reload_InvalidDir(t *testing.T) {
	// 空目录且没有 Global.json 时，Parse 会失败，Reload 不会 Store，getAll 仍为空
	emptyDir := t.TempDir()
	m := NewManager(emptyDir, false, Parse)
	m.Reload()
	all := m.getAll()
	if len(all) != 0 {
		t.Errorf("Reload with dir missing Global.json should not store new data, got len(all)=%d", len(all))
	}
}

func TestManager_getAll_InitialEmpty(t *testing.T) {
	m := NewManager("", false, Parse)
	all := m.getAll()
	if all == nil {
		t.Fatal("getAll() should not return nil")
	}
	if len(all) != 0 {
		t.Errorf("initial getAll() len = %d, want 0", len(all))
	}
}

// keys 返回 map[string]any 的 key 列表，用于错误信息
func keys(m map[string]any) []string {
	if m == nil {
		return nil
	}
	k := make([]string, 0, len(m))
	for s := range m {
		k = append(k, s)
	}
	return k
}

// ---------------------------------------------------------------------------
// 集成测试：使用真实配置目录校验（需在项目根目录执行或指定 CONFIG_DIR）
// 用法: go test ./core/config -run TestLoadRealConfigDir -v
// 或:   CONFIG_DIR=./core/config/json go test ./core/config -run TestLoadRealConfigDir -v
// ---------------------------------------------------------------------------

func TestLoadRealConfigDir(t *testing.T) {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		// 默认尝试同包下的 json 目录（与 config_test.go 同目录）
		_, filename, _, _ := runtime.Caller(0)
		configDir = filepath.Join(filepath.Dir(filename), "json")
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Skipf("skip integration test: config dir not found: %s (set CONFIG_DIR to run)", configDir)
		return
	}
	m := NewManager(configDir, false, Parse)
	m.Reload()
	all := m.getAll()
	if len(all) == 0 {
		t.Fatal("real config dir load failed: getAll() is empty")
	}
	if _, ok := all["Global"]; !ok {
		t.Errorf("real config missing Global, keys: %v", keys(all))
	}
	t.Logf("loaded %d config keys from %s", len(all), configDir)
}
