package main

import (
	"flag"
	"fmt"
	"os"

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
	for _, ib := range cfg.Inbounds {
		if p, ok := ib["listen_port"].(float64); ok && int(p) == port {
			proto = ib["type"].(string)
			break
		}
	}
	if proto == "" {
		for _, ep := range cfg.Endpoints {
			if p, ok := ep["listen_port"].(float64); ok && int(p) == port {
				proto = ep["type"].(string)
				break
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
			if p, ok := ib["listen_port"].(float64); ok && int(p) == port {
				sp = ib["password"].(string)
				break
			}
		}
		user := map[string]interface{}{"name": remark, "password": up}
		config.AppendInboundUser(cfg, port, user)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintSS(ip, port, remark, sp, up)

	case "vmess":
		uuid := generator.GenerateUUID()
		user := map[string]interface{}{"name": remark, "uuid": uuid, "alterId": 0}
		config.AppendInboundUser(cfg, port, user)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintVMess(ip, port, remark, uuid)

	case "hysteria2":
		pw, _ := generator.GeneratePassword()
		user := map[string]interface{}{"name": remark, "password": pw}
		config.AppendInboundUser(cfg, port, user)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintHysteria2(ip, port, remark, pw)

	case "wireguard":
		var sPriv string
		var peerCount int
		for _, ep := range cfg.Endpoints {
			if p, ok := ep["listen_port"].(float64); ok && int(p) == port {
				sPriv = ep["private_key"].(string)
				if peers, ok := ep["peers"].([]interface{}); ok {
					peerCount = len(peers)
				}
				break
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

		peer := map[string]interface{}{
			"public_key":     cPub,
			"pre_shared_key": psk,
			"allowed_ips":    []string{userIP + "/32", fmt.Sprintf("fd00::%d/128", userIdx)},
			"comment":        remark,
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
