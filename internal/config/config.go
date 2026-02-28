package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultConfigPath = "/etc/sing-box/config.json"
	DefaultConfigDir  = "/etc/sing-box"
)

// SingBoxConfig 定义我们要读写的 config.json 结构
// 因为我们只关心我们要追加的字段，其它字段可以使用 map[string]interface{} 或 json.RawMessage 原样保留
type SingBoxConfig struct {
	Log       map[string]interface{}   `json:"log,omitempty"`
	DNS       map[string]interface{}   `json:"dns,omitempty"`
	Inbounds  []map[string]interface{} `json:"inbounds"`
	Endpoints []map[string]interface{} `json:"endpoints"`
	Outbounds []map[string]interface{} `json:"outbounds"`
	Route     map[string]interface{}   `json:"route,omitempty"`
}

// ReadConfig 读取配置文件
func ReadConfig(path string) (*SingBoxConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg SingBoxConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// WriteConfig 回写配置
func WriteConfig(path string, cfg *SingBoxConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// InitBaseConfig 初始化基础骨架配置
func InitBaseConfig(path string) (*SingBoxConfig, error) {
	if _, err := os.Stat(path); err == nil {
		return ReadConfig(path)
	}

	cfg := &SingBoxConfig{
		Log: map[string]interface{}{"level": "info"},
		DNS: map[string]interface{}{
			"servers": []interface{}{
				map[string]string{
					"type":   "tls",
					"server": "8.8.8.8",
				},
			},
		},
		Inbounds:  []map[string]interface{}{},
		Endpoints: []map[string]interface{}{},
		Outbounds: []map[string]interface{}{
			{"type": "direct"},
		},
		Route: map[string]interface{}{
			"rules": []interface{}{
				map[string]string{
					"protocol": "dns",
					"action":   "hijack-dns",
				},
			},
		},
	}
	err := WriteConfig(path, cfg)
	return cfg, err
}

// CheckPortConflict 检查给定的端口是否被当前任何 inbound/endpoint 占用
func CheckPortConflict(cfg *SingBoxConfig, port int) bool {
	for _, ib := range cfg.Inbounds {
		if p, ok := ib["listen_port"].(float64); ok && int(p) == port {
			return true
		}
	}
	for _, ep := range cfg.Endpoints {
		if p, ok := ep["listen_port"].(float64); ok && int(p) == port {
			return true
		}
	}
	return false
}

// EnsureDirectInbound 如果 endpoint 使用了 wg，需要确保有一个空的 direct inbound 兜底
func EnsureDirectInbound(cfg *SingBoxConfig, fallbackPort int) {
	for _, ib := range cfg.Inbounds {
		if t, ok := ib["type"].(string); ok && t == "direct" {
			return
		}
	}
	cfg.Inbounds = append(cfg.Inbounds, map[string]interface{}{
		"type":        "direct",
		"tag":         "direct-in",
		"listen":      "::",
		"listen_port": fallbackPort,
	})
}
