package config

// 轻量级 sing-box 类型注册表
// 只映射本工具支持的协议类型，避免引入 include 包的全量依赖。

import (
	"context"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/service"
)

// singBoxContext 返回注册了所有必要类型注册表的 context，
// 使 sing-box 的 json 包能正确解析 inbound/endpoint/dns 等 typed union 字段。
func singBoxContext() context.Context {
	ctx := context.Background()
	ctx = service.ContextWith[option.InboundOptionsRegistry](ctx, new(minimalInboundRegistry))
	ctx = service.ContextWith[option.EndpointOptionsRegistry](ctx, new(minimalEndpointRegistry))
	ctx = service.ContextWith[option.OutboundOptionsRegistry](ctx, new(minimalOutboundRegistry))
	ctx = service.ContextWith[option.DNSTransportOptionsRegistry](ctx, new(minimalDNSRegistry))
	ctx = service.ContextWith[option.ServiceOptionsRegistry](ctx, new(minimalServiceRegistry))
	return ctx
}

// ---- Inbound ----

type minimalInboundRegistry struct{}

func (r *minimalInboundRegistry) CreateOptions(t string) (any, bool) {
	switch t {
	case "shadowsocks":
		return new(option.ShadowsocksInboundOptions), true
	case "vmess":
		return new(option.VMessInboundOptions), true
	case "hysteria2":
		return new(option.Hysteria2InboundOptions), true
	case "hysteria":
		return new(option.HysteriaInboundOptions), true
	case "trojan":
		return new(option.TrojanInboundOptions), true
	case "vless":
		return new(option.VLESSInboundOptions), true
	case "shadowtls":
		return new(option.ShadowTLSInboundOptions), true
	case "tuic":
		return new(option.TUICInboundOptions), true
	case "naive":
		return new(option.NaiveInboundOptions), true
	case "tun":
		return new(option.TunInboundOptions), true
	case "redirect":
		return new(option.RedirectInboundOptions), true
	case "tproxy":
		return new(option.TProxyInboundOptions), true
	case "socks":
		return new(option.SocksInboundOptions), true
	case "http", "mixed":
		return new(option.HTTPMixedInboundOptions), true
	case "direct":
		return new(option.DirectInboundOptions), true
	case "anytls":
		return new(option.AnyTLSInboundOptions), true
	default:
		return nil, false
	}
}

// ---- Endpoint ----

type minimalEndpointRegistry struct{}

func (r *minimalEndpointRegistry) CreateOptions(t string) (any, bool) {
	switch t {
	case "wireguard":
		return new(option.WireGuardEndpointOptions), true
	default:
		return nil, false
	}
}

// ---- Outbound ----

type minimalOutboundRegistry struct{}

func (r *minimalOutboundRegistry) CreateOptions(t string) (any, bool) {
	switch t {
	case "direct":
		return new(option.DirectOutboundOptions), true
	case "shadowsocks":
		return new(option.ShadowsocksOutboundOptions), true
	case "vmess":
		return new(option.VMessOutboundOptions), true
	case "hysteria2":
		return new(option.Hysteria2OutboundOptions), true
	case "block", "dns":
		return &struct{}{}, true
	default:
		// 未知出站类型：返回空结构体，保留 type/tag 字段
		return &struct{}{}, true
	}
}

// ---- DNS Transport ----

type minimalDNSRegistry struct{}

func (r *minimalDNSRegistry) CreateOptions(t string) (any, bool) {
	switch t {
	case "udp", "tcp":
		return new(option.RemoteDNSServerOptions), true
	case "tls", "quic":
		return new(option.RemoteTLSDNSServerOptions), true
	case "https", "h3":
		return new(option.RemoteHTTPSDNSServerOptions), true
	case "local":
		return new(option.LocalDNSServerOptions), true
	case "fakeip":
		return new(option.FakeIPDNSServerOptions), true
	case "dhcp":
		return new(option.DHCPDNSServerOptions), true
	case "hosts":
		return new(option.HostsDNSServerOptions), true
	default:
		return nil, false
	}
}

// ---- Service ----

type minimalServiceRegistry struct{}

func (r *minimalServiceRegistry) CreateOptions(t string) (any, bool) {
	return nil, false
}
