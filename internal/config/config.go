package config

import (
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

const (
	DefaultConfigPath = "/etc/sing-box/config.json"
	DefaultConfigDir  = "/etc/sing-box"
)

// SingBoxConfig 使用 sing-box 官方的配置类型
type SingBoxConfig = option.Options

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
		Log: &option.LogOptions{
			Level: "info",
		},
		Inbounds:  []option.Inbound{},
		Endpoints: []option.Endpoint{},
		Outbounds: []option.Outbound{
			{
				Type: "direct",
				Tag:  "direct",
			},
		},
		Route: &option.RouteOptions{
			AutoDetectInterface: true,
		},
	}
	err := WriteConfig(path, cfg)
	return cfg, err
}

// CheckPortConflict 检查给定的端口是否被当前任何 inbound/endpoint 占用
func CheckPortConflict(cfg *SingBoxConfig, port int) bool {
	// 检查 inbounds
	for _, ib := range cfg.Inbounds {
		if opts, ok := ib.Options.(interface{ GetListenOptions() *option.ListenOptions }); ok {
			if listenOpts := opts.GetListenOptions(); listenOpts != nil && listenOpts.ListenPort == uint16(port) {
				return true
			}
		}
	}
	// 检查 endpoints (WireGuard)
	for _, ep := range cfg.Endpoints {
		if ep.Type == "wireguard" {
			if wgOpts, ok := ep.Options.(*option.WireGuardEndpointOptions); ok {
				if wgOpts.ListenPort == uint16(port) {
					return true
				}
			}
		}
	}
	return false
}

// EnsureDirectInbound 如果 endpoint 使用了 wg，需要确保有一个空的 direct inbound 兜底
func EnsureDirectInbound(cfg *SingBoxConfig, fallbackPort int) {
	// 检查是否已存在 direct inbound
	for _, ib := range cfg.Inbounds {
		if ib.Type == "direct" {
			return
		}
	}
	
	// 添加 direct inbound
	listenAddr := "::"
	addr, _ := netip.ParseAddr(listenAddr)
	badAddr := (*badoption.Addr)(&addr)
	
	cfg.Inbounds = append(cfg.Inbounds, option.Inbound{
		Type:    "direct",
		Tag:     "direct-in",
		Options: &option.DirectInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: uint16(fallbackPort),
			},
		},
	})
}
