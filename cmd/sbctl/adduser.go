package main

import (
	"flag"
	"fmt"
	"net/netip"
	"os"

	"github.com/sagernet/sing-box/option"
	"github.com/sail-tunnel/sbctl/internal/config"
	"github.com/sail-tunnel/sbctl/internal/generator"
	"github.com/sail-tunnel/sbctl/internal/protocols"
	"github.com/sail-tunnel/sbctl/internal/system"
)

// =========================================================
// add-user 命令
// =========================================================
func cmdAddUser() {
	if !system.CheckRoot() {
		fmt.Println("[ERROR] 请使用 root 权限运行")
		os.Exit(1)
	}

	addCmd := flag.NewFlagSet("add-user", flag.ExitOnError)
	var remark string
	var port int

	addCmd.StringVar(&remark, "r", "", "备注名")
	addCmd.StringVar(&remark, "remark", "", "备注名")
	addCmd.IntVar(&port, "p", 0, "端口")
	addCmd.IntVar(&port, "port", 0, "端口")

	addCmd.Parse(os.Args[2:])

	if port == 0 || remark == "" {
		fmt.Println("[ERROR] 缺少参数！用法: sbctl add-user -p <监听端口> -r <用户备注>")
		os.Exit(1)
	}

	cfg, err := config.ReadConfig(config.DefaultConfigPath)
	if err != nil {
		fmt.Println("[ERROR] 读取配置文件失败，请先用 'sbctl init' 安装协议")
		os.Exit(1)
	}

	var proto string
	// 检查 inbounds
	for _, ib := range cfg.Inbounds {
		var listenPort uint16
		switch opts := ib.Options.(type) {
		case *option.ShadowsocksInboundOptions:
			listenPort = opts.ListenPort
		case *option.VMessInboundOptions:
			listenPort = opts.ListenPort
		case *option.Hysteria2InboundOptions:
			listenPort = opts.ListenPort
		}
		if listenPort == uint16(port) {
			proto = ib.Type
			break
		}
	}
	// 检查 endpoints
	if proto == "" {
		for _, ep := range cfg.Endpoints {
			if ep.Type == "wireguard" {
				if wgOpts, ok := ep.Options.(*option.WireGuardEndpointOptions); ok {
					if wgOpts.ListenPort == uint16(port) {
						proto = ep.Type
						break
					}
				}
			}
		}
	}

	if proto == "" {
		fmt.Printf("[ERROR] 未在配置中找到端口 %d，请先用 'sbctl init' 安装该端口\n", port)
		os.Exit(1)
	}

	fmt.Printf("[INFO] 在端口 %d (%s) 追加新客户端: %s ...\n", port, proto, remark)
	ip := generator.GetPublicIP()

	switch proto {
	case "shadowsocks":
		up, _ := generator.GeneratePassword()
		var sp string
		for _, ib := range cfg.Inbounds {
			if opts, ok := ib.Options.(*option.ShadowsocksInboundOptions); ok {
				if opts.ListenPort == uint16(port) {
					sp = opts.Password
					break
				}
			}
		}
		user := option.ShadowsocksUser{Name: remark, Password: up}
		config.AppendInboundUser(cfg, port, user)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintSS(ip, port, remark, sp, up)

	case "vmess":
		uuid := generator.GenerateUUID()
		user := option.VMessUser{Name: remark, UUID: uuid}
		config.AppendInboundUser(cfg, port, user)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintVMess(ip, port, remark, uuid)

	case "hysteria2":
		pw, _ := generator.GeneratePassword()
		user := option.Hysteria2User{Name: remark, Password: pw}
		config.AppendInboundUser(cfg, port, user)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintHysteria2(ip, port, remark, pw)

	case "wireguard":
		var sPriv string
		var peerCount int
		for _, ep := range cfg.Endpoints {
			if ep.Type == "wireguard" {
				if wgOpts, ok := ep.Options.(*option.WireGuardEndpointOptions); ok {
					if wgOpts.ListenPort == uint16(port) {
						sPriv = wgOpts.PrivateKey
						peerCount = len(wgOpts.Peers)
						break
					}
				}
			}
		}

		sPub, err := generator.GetWGPublicKeyFromString(sPriv)
		if err != nil {
			fmt.Printf("[ERROR] 派生服务端公钥时遇到错误: %v\n", err)
			os.Exit(1)
		}

		cPriv, cPub, _ := generator.GenerateWGKeyPair()
		psk, _ := generator.GeneratePassword()
		userIdx := peerCount + 2
		userIP := fmt.Sprintf("10.0.0.%d", userIdx)

		peerAddr1, _ := netip.ParsePrefix(fmt.Sprintf("%s/32", userIP))
		peerAddr2, _ := netip.ParsePrefix(fmt.Sprintf("fd00::%d/128", userIdx))

		peer := option.WireGuardPeer{
			PublicKey:    cPub,
			PreSharedKey: psk,
			AllowedIPs:   []netip.Prefix{peerAddr1, peerAddr2},
		}

		if err := config.AppendEndpointPeer(cfg, port, peer); err != nil {
			fmt.Printf("[ERROR] 追加 Peer 失败: %v\n", err)
			os.Exit(1)
		}

		if err := config.WriteConfig(config.DefaultConfigPath, cfg); err != nil {
			fmt.Printf("[ERROR] 写入配置失败: %v\n", err)
			os.Exit(1)
		}

		protocols.PrintWireGuard(ip, port, remark, cPriv, sPub, psk, userIP)

	default:
		fmt.Println("[ERROR] 未知或不支持的协议:", proto)
		os.Exit(1)
	}

	system.RestartService()
	fmt.Println("[INFO] 新用户追加完成！服务已重启并生效。")
}
