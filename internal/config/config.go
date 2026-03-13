package config

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sync"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badoption"
)

const (
	DefaultConfigDir  = "/etc/sing-box"
	DefaultConfigPath = DefaultConfigDir + "/config.json"
)

// 用于测试的证书目录覆盖
var CertsDirOverride string

// GetCertsDir 返回证书目录，测试时可以通过 CertsDirOverride 覆盖
func GetCertsDir() string {
	if CertsDirOverride != "" {
		return CertsDirOverride
	}
	return DefaultConfigDir + "/certs"
}

// SingBoxConfig 使用 sing-box 官方的配置类型
type SingBoxConfig = option.Options

// rawDNSStore 保存无法被 sing-box json 解析的 DNS 原始字节，以便 WriteConfig 时还原
var rawDNSStore sync.Map // *SingBoxConfig → []byte

// ReadConfig 读取配置文件
func ReadConfig(path string) (*SingBoxConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg SingBoxConfig

	// 优先用 sing-box 的 context-aware 解析（处理协议类型注册）
	if err := json.UnmarshalContext(context.Background(), data, &cfg); err == nil {
		return &cfg, nil
	}

	// Fallback：sing-box json 解析失败（通常是 DNS transport 类型未注册）
	// 用标准库把 dns 字段提取为原始 JSON，其余部分继续用 sing-box json 解析
	var rawMap map[string]stdjson.RawMessage
	if err := stdjson.Unmarshal(data, &rawMap); err != nil {
		return nil, fmt.Errorf("配置文件格式无效: %w", err)
	}

	dnsRaw := rawMap["dns"]
	delete(rawMap, "dns")

	stripped, err := stdjson.Marshal(rawMap)
	if err != nil {
		return nil, err
	}

	if err := json.UnmarshalContext(context.Background(), stripped, &cfg); err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %w", err)
	}

	// 保存原始 DNS JSON，WriteConfig 时原样写回
	if dnsRaw != nil {
		rawDNSStore.Store(&cfg, []byte(dnsRaw))
	}

	return &cfg, nil
}

// WriteConfig 回写配置
func WriteConfig(path string, cfg *SingBoxConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := json.NewEncoderContext(context.Background(), &buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	// 如果读取时 DNS 无法被解析，将原始 DNS JSON 注入回输出
	if dnsRaw, ok := rawDNSStore.Load(cfg); ok {
		var rawMap map[string]stdjson.RawMessage
		if err := stdjson.Unmarshal(buf.Bytes(), &rawMap); err == nil {
			rawMap["dns"] = stdjson.RawMessage(dnsRaw.([]byte))
			if out, err := stdjson.MarshalIndent(rawMap, "", "  "); err == nil {
				return os.WriteFile(path, append(out, '\n'), 0644)
			}
		}
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// InitBaseConfig 初始化基础骨架配置
func InitBaseConfig(path string) (*SingBoxConfig, error) {
	if _, err := os.Stat(path); err == nil {
		// 文件存在，读取并确保结构完整
		cfg, err := ReadConfig(path)
		if err != nil {
			return nil, err
		}
		// 确保配置结构完整
		ensureCompleteConfig(cfg)
		// 写回配置以补全缺失的字段
		if err := WriteConfig(path, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	return createDefaultConfig(path)
}

// ResetConfig 强制重置配置为默认配置（即使文件已存在）
func ResetConfig(path string) (*SingBoxConfig, error) {
	return createDefaultConfig(path)
}

// ensureCompleteConfig 确保配置结构完整，补全缺失的字段
func ensureCompleteConfig(cfg *SingBoxConfig) {
	// 确保 Log 配置存在
	if cfg.Log == nil {
		cfg.Log = &option.LogOptions{
			Level: "info",
		}
	}

	// 确保 Inbounds 和 Endpoints 不为 nil
	if cfg.Inbounds == nil {
		cfg.Inbounds = []option.Inbound{}
	}
	if cfg.Endpoints == nil {
		cfg.Endpoints = []option.Endpoint{}
	}

	// 确保至少有一个 direct outbound
	if cfg.Outbounds == nil || len(cfg.Outbounds) == 0 {
		cfg.Outbounds = []option.Outbound{
			{
				Type: "direct",
				Tag:  "direct",
			},
		}
	} else {
		// 检查是否已存在 direct outbound
		hasDirectOutbound := false
		for _, outbound := range cfg.Outbounds {
			if outbound.Type == "direct" {
				hasDirectOutbound = true
				break
			}
		}
		if !hasDirectOutbound {
			cfg.Outbounds = append(cfg.Outbounds, option.Outbound{
				Type:    "direct",
				Tag:     "direct",
				Options: &option.DirectOutboundOptions{},
			})
		}
	}

	// 确保 Route 配置存在并包含必要的规则
	if cfg.Route == nil {
		cfg.Route = &option.RouteOptions{
			AutoDetectInterface: true,
			Rules: []option.Rule{
				{
					Type: "default",
					DefaultOptions: option.DefaultRule{
						RuleAction: option.RuleAction{
							Action: "sniff",
						},
					},
				},
				{
					Type: "default",
					DefaultOptions: option.DefaultRule{
						RawDefaultRule: option.RawDefaultRule{
							Protocol: badoption.Listable[string]{"dns"},
						},
						RuleAction: option.RuleAction{
							Action: "hijack-dns",
						},
					},
				},
			},
		}
	} else {
		// 确保 AutoDetectInterface 开启
		cfg.Route.AutoDetectInterface = true

		// 检查并添加 sniff 和 hijack-dns 规则（如果不存在）
		if cfg.Route.Rules == nil {
			cfg.Route.Rules = []option.Rule{}
		}

		hasSniffRule := false
		hasHijackDNSRule := false
		for _, rule := range cfg.Route.Rules {
			if rule.Type == "default" {
				if rule.DefaultOptions.RuleAction.Action == "sniff" {
					hasSniffRule = true
				}
				if rule.DefaultOptions.RuleAction.Action == "hijack-dns" {
					hasHijackDNSRule = true
				}
			}
		}

		// 将必要的规则插入到开头
		newRules := []option.Rule{}
		if !hasSniffRule {
			newRules = append(newRules, option.Rule{
				Type: "default",
				DefaultOptions: option.DefaultRule{
					RuleAction: option.RuleAction{
						Action: "sniff",
					},
				},
			})
		}
		if !hasHijackDNSRule {
			newRules = append(newRules, option.Rule{
				Type: "default",
				DefaultOptions: option.DefaultRule{
					RawDefaultRule: option.RawDefaultRule{
						Protocol: badoption.Listable[string]{"dns"},
					},
					RuleAction: option.RuleAction{
						Action: "hijack-dns",
					},
				},
			})
		}
		cfg.Route.Rules = append(newRules, cfg.Route.Rules...)
	}

	// 确保 DNS 配置存在
	if cfg.DNS == nil {
		cfg.DNS = &option.DNSOptions{
			RawDNSOptions: option.RawDNSOptions{
				Servers: []option.DNSServerOptions{
					{
						Tag:  "google",
						Type: "udp",
						Options: &option.RemoteDNSServerOptions{
							DNSServerAddressOptions: option.DNSServerAddressOptions{
								Server: "8.8.8.8",
							},
						},
					},
				},
				Rules: []option.DNSRule{
					{
						Type: "default",
						DefaultOptions: option.DefaultDNSRule{
							DNSRuleAction: option.DNSRuleAction{
								Action: "route",
								RouteOptions: option.DNSRouteActionOptions{
									Server: "google",
								},
							},
						},
					},
				},
			},
		}
	}
}

// createDefaultConfig 创建默认配置
func createDefaultConfig(path string) (*SingBoxConfig, error) {
	cfg := &SingBoxConfig{
		Log: &option.LogOptions{
			Level: "info",
		},
		Inbounds:  []option.Inbound{},
		Endpoints: []option.Endpoint{},
		Outbounds: []option.Outbound{
			{
				Type:    "direct",
				Tag:     "direct",
				Options: &option.DirectOutboundOptions{},
			},
		},
		Route: &option.RouteOptions{
			AutoDetectInterface: true,
			Rules: []option.Rule{
				{
					Type: "default",
					DefaultOptions: option.DefaultRule{
						RuleAction: option.RuleAction{
							Action: "sniff",
						},
					},
				},
				{
					Type: "default",
					DefaultOptions: option.DefaultRule{
						RawDefaultRule: option.RawDefaultRule{
							Protocol: badoption.Listable[string]{"dns"},
						},
						RuleAction: option.RuleAction{
							Action: "hijack-dns",
						},
					},
				},
			},
		},
		DNS: &option.DNSOptions{
			RawDNSOptions: option.RawDNSOptions{
				Servers: []option.DNSServerOptions{
					{
						Tag:  "google",
						Type: "udp",
						Options: &option.RemoteDNSServerOptions{
							DNSServerAddressOptions: option.DNSServerAddressOptions{
								Server: "8.8.8.8",
							},
						},
					},
				},
				Rules: []option.DNSRule{
					{
						Type: "default",
						DefaultOptions: option.DefaultDNSRule{
							DNSRuleAction: option.DNSRuleAction{
								Action: "route",
								RouteOptions: option.DNSRouteActionOptions{
									Server: "google",
								},
							},
						},
					},
				},
			},
		},
	}
	err := WriteConfig(path, cfg)
	return cfg, err
}

// CheckPortConflict 检查给定的端口是否被当前任何 inbound/endpoint 占用
func CheckPortConflict(cfg *SingBoxConfig, port int) bool {
	// 检查 inbounds - 使用类型断言检查各种类型的 inbound options
	for _, ib := range cfg.Inbounds {
		var listenPort uint16

		switch opts := ib.Options.(type) {
		case *option.ShadowsocksInboundOptions:
			listenPort = opts.ListenPort
		case *option.VMessInboundOptions:
			listenPort = opts.ListenPort
		case *option.Hysteria2InboundOptions:
			listenPort = opts.ListenPort
		case *option.DirectInboundOptions:
			listenPort = opts.ListenPort
		case interface{ GetListenOptions() *option.ListenOptions }:
			if listenOpts := opts.GetListenOptions(); listenOpts != nil {
				listenPort = listenOpts.ListenPort
			}
		}

		if listenPort == uint16(port) {
			return true
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

// HasMeaningfulConfig 检查配置是否包含有意义的节点配置
// 如果配置文件不存在或只包含默认的空配置，返回 false
func HasMeaningfulConfig(path string) bool {
	cfg, err := ReadConfig(path)
	if err != nil {
		return false
	}

	// 检查是否有实际的节点配置
	// 注意：我们不计算默认的 direct outbound
	meaningfulInbounds := len(cfg.Inbounds) > 0
	meaningfulEndpoints := len(cfg.Endpoints) > 0

	return meaningfulInbounds || meaningfulEndpoints
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
		Type: "direct",
		Tag:  "direct-in",
		Options: &option.DirectInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: uint16(fallbackPort),
			},
		},
	})
}
