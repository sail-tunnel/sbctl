package config

import (
	"encoding/json"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

// TestCreateDefaultConfig 测试创建默认配置
func TestCreateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg, err := createDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("createDefaultConfig failed: %v", err)
	}

	// 验证配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// 验证基本结构
	if cfg.Log == nil {
		t.Error("Log config is nil")
	}
	if cfg.DNS == nil {
		t.Error("DNS config is nil")
	}
	if cfg.Route == nil {
		t.Error("Route config is nil")
	}
	if len(cfg.Outbounds) == 0 {
		t.Error("Outbounds is empty")
	}

	// 验证必要的路由规则
	if cfg.Route.Rules == nil || len(cfg.Route.Rules) < 2 {
		t.Error("Route rules are missing or incomplete")
	}

	// 使用 sing-box check 验证配置格式
	verifySingBoxConfig(t, configPath)
}

// TestInitBaseConfig 测试初始化基础配置
func TestInitBaseConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 第一次调用应该创建新配置
	cfg, err := InitBaseConfig(configPath)
	if err != nil {
		t.Fatalf("InitBaseConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("config is nil")
	}

	// 验证配置结构完整
	if cfg.Log == nil || cfg.DNS == nil || cfg.Route == nil {
		t.Error("Config structure is incomplete")
	}

	// 验证配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	verifySingBoxConfig(t, configPath)
}

// TestResetConfig 测试重置配置
func TestResetConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 先创建一个配置
	cfg, err := InitBaseConfig(configPath)
	if err != nil {
		t.Fatalf("InitBaseConfig failed: %v", err)
	}

	// 手动添加一个 inbound
	listenAddr, _ := netip.ParseAddr("::")
	badAddr := (*badoption.Addr)(&listenAddr)
	cfg.Inbounds = append(cfg.Inbounds, option.Inbound{
		Type: "shadowsocks",
		Tag:  "ss-in-12345",
		Options: &option.ShadowsocksInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: 12345,
			},
			Method:   "2022-blake3-aes-256-gcm",
			Password: "test-password",
		},
	})

	if err := WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	// 验证有 inbound
	if len(cfg.Inbounds) == 0 {
		t.Fatal("Expected inbound to be added")
	}

	// 重置配置
	resetCfg, err := ResetConfig(configPath)
	if err != nil {
		t.Fatalf("ResetConfig failed: %v", err)
	}

	// 验证配置被重置
	if len(resetCfg.Inbounds) != 0 {
		t.Error("Expected inbounds to be empty after reset")
	}

	verifySingBoxConfig(t, configPath)
}

// TestWriteConfig 测试写入配置
func TestWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 创建一个配置
	cfg, err := InitBaseConfig(configPath)
	if err != nil {
		t.Fatalf("InitBaseConfig failed: %v", err)
	}

	// 手动添加一个 inbound
	listenAddr, _ := netip.ParseAddr("::")
	badAddr := (*badoption.Addr)(&listenAddr)
	cfg.Inbounds = append(cfg.Inbounds, option.Inbound{
		Type: "shadowsocks",
		Tag:  "ss-in-23456",
		Options: &option.ShadowsocksInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: 23456,
			},
			Method:   "2022-blake3-aes-256-gcm",
			Password: "test-server-pass",
			Users: []option.ShadowsocksUser{
				{
					Name:     "user1",
					Password: "test-user-pass",
				},
			},
			Multiplex: &option.InboundMultiplexOptions{
				Enabled: true,
				Padding: true,
			},
		},
	})

	// 写入配置
	if err := WriteConfig(configPath, cfg); err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	// 验证文件存在且可以被 sing-box check 验证
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// 读取文件内容验证关键字段存在
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	content := string(data)
	// 验证关键字段存在
	requiredFields := []string{
		`"type": "shadowsocks"`,
		`"listen": "::"`,
		`"listen_port": 23456`,
		`"method": "2022-blake3-aes-256-gcm"`,
		`"password": "test-server-pass"`,
		`"users"`,
		`"server": "8.8.8.8"`,    // DNS server
		`"action": "sniff"`,      // Route action
		`"action": "hijack-dns"`, // DNS hijack action
	}

	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("Config missing required field: %s", field)
		}
	}

	verifySingBoxConfig(t, configPath)
}

// TestEnsureCompleteConfig 测试确保配置完整
func TestEnsureCompleteConfig(t *testing.T) {
	// 创建一个不完整的配置
	cfg := &SingBoxConfig{}

	ensureCompleteConfig(cfg)

	// 验证所有必要字段都已添加
	if cfg.Log == nil {
		t.Error("Log config was not added")
	}
	if cfg.DNS == nil {
		t.Error("DNS config was not added")
	}
	if cfg.Route == nil {
		t.Error("Route config was not added")
	}
	if len(cfg.Outbounds) == 0 {
		t.Error("Outbounds were not added")
	}
	if cfg.Inbounds == nil {
		t.Error("Inbounds should be initialized")
	}
	if cfg.Endpoints == nil {
		t.Error("Endpoints should be initialized")
	}

	// 验证路由规则
	if len(cfg.Route.Rules) < 2 {
		t.Error("Route rules are incomplete")
	}

	// 验证 sniff 和 hijack-dns 规则存在
	hasSniff := false
	hasHijackDNS := false
	for _, rule := range cfg.Route.Rules {
		if rule.Type == "default" {
			if rule.DefaultOptions.RuleAction.Action == "sniff" {
				hasSniff = true
			}
			if rule.DefaultOptions.RuleAction.Action == "hijack-dns" {
				hasHijackDNS = true
			}
		}
	}

	if !hasSniff {
		t.Error("Sniff rule was not added")
	}
	if !hasHijackDNS {
		t.Error("Hijack-DNS rule was not added")
	}
}

// TestCheckPortConflict 测试端口冲突检查
func TestCheckPortConflict(t *testing.T) {
	cfg := &SingBoxConfig{}
	ensureCompleteConfig(cfg)

	// 手动添加一个 inbound
	listenAddr, _ := netip.ParseAddr("::")
	badAddr := (*badoption.Addr)(&listenAddr)
	cfg.Inbounds = append(cfg.Inbounds, option.Inbound{
		Type: "shadowsocks",
		Tag:  "ss-in-8888",
		Options: &option.ShadowsocksInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: 8888,
			},
			Method:   "2022-blake3-aes-256-gcm",
			Password: "test-password",
		},
	})

	// 检查同一个端口应该冲突
	if !CheckPortConflict(cfg, 8888) {
		t.Error("Expected port conflict for 8888")
	}

	// 检查不同端口不应该冲突
	if CheckPortConflict(cfg, 9999) {
		t.Error("Expected no port conflict for 9999")
	}
}

// verifySingBoxConfig 验证配置文件
func verifySingBoxConfig(t *testing.T, configPath string) {
	t.Helper()

	// 1. 基本验证：确保文件是有效的 JSON
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("Failed to read config file: %v", err)
		return
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		t.Errorf("Config is not valid JSON: %v", err)
		return
	}

	// 2. 验证必要的顶层字段存在
	requiredFields := []string{"log", "dns", "outbounds", "route"}
	for _, field := range requiredFields {
		if _, ok := rawConfig[field]; !ok {
			t.Errorf("Config missing required field: %s", field)
		}
	}

	// 3. 如果 sing-box 命令可用，使用它进行完整验证
	if singBoxPath, err := exec.LookPath("sing-box"); err == nil {
		cmd := exec.Command(singBoxPath, "check", "-c", configPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("sing-box check failed: %v\nOutput: %s", err, string(output))
			return
		}
		t.Logf("sing-box check passed for %s", configPath)
	} else {
		t.Logf("sing-box command not available, performed basic validation only for %s", configPath)
	}
}
