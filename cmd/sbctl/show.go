package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sagernet/sing-box/option"
	"github.com/sail-tunnel/sbctl/internal/config"
	"github.com/sail-tunnel/sbctl/internal/generator"
	"github.com/sail-tunnel/sbctl/internal/protocols"
)

// =========================================================
// show 命令
// =========================================================
func cmdShow() {
	showCmd := flag.NewFlagSet("show", flag.ExitOnError)
	var targetPort int

	showCmd.IntVar(&targetPort, "p", 0, "端口")
	showCmd.IntVar(&targetPort, "port", 0, "端口")

	showCmd.Parse(os.Args[2:])

	cfg, err := config.ReadConfig(config.GetConfigPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("[ERROR] 配置文件不存在：%s\n请先运行 'sbctl init' 进行安装。\n", config.DefaultConfigPath)
		} else {
			fmt.Printf("[ERROR] 配置文件解析失败：%v\n", err)
		}
		os.Exit(1)
	}

	ip := generator.GetPublicIP()
	found := false

	fmt.Println("\n=================================================")
	fmt.Printf(" 当前服务端 IP : %s\n", ip)
	fmt.Println("=================================================")

	// 遍历 inbounds
	for _, ib := range cfg.Inbounds {
		var port uint16

		switch opts := ib.Options.(type) {
		case *option.ShadowsocksInboundOptions:
			port = opts.ListenPort
		case *option.VMessInboundOptions:
			port = opts.ListenPort
		case *option.Hysteria2InboundOptions:
			port = opts.ListenPort
		default:
			continue
		}

		if targetPort != 0 && targetPort != int(port) {
			continue
		}

		switch ib.Type {
		case "shadowsocks":
			ssOpts := ib.Options.(*option.ShadowsocksInboundOptions)
			for _, user := range ssOpts.Users {
				protocols.PrintSS(ip, int(port), user.Name, ssOpts.Password, user.Password)
				found = true
			}

		case "vmess":
			vmOpts := ib.Options.(*option.VMessInboundOptions)
			for _, user := range vmOpts.Users {
				protocols.PrintVMess(ip, int(port), user.Name, user.UUID)
				found = true
			}

		case "hysteria2":
			hy2Opts := ib.Options.(*option.Hysteria2InboundOptions)
			for _, user := range hy2Opts.Users {
				protocols.PrintHysteria2(ip, int(port), user.Name, user.Password)
				found = true
			}
		}
	}

	// 遍历 endpoints
	for _, ep := range cfg.Endpoints {
		if ep.Type != "wireguard" {
			continue
		}

		wgOpts, ok := ep.Options.(*option.WireGuardEndpointOptions)
		if !ok {
			continue
		}

		port := wgOpts.ListenPort
		if targetPort != 0 && targetPort != int(port) {
			continue
		}

		for i, peer := range wgOpts.Peers {
			remark := fmt.Sprintf("Peer-%d", i)

			fmt.Printf("\n========== WireGuard ==========\n")
			fmt.Printf("服务器: %s  端口: %d  备注: %s\n", ip, port, remark)
			fmt.Println("[WARN] 由于 WireGuard 私钥仅在生成时可知，无法重新完整展示客户端配置文件。")
			fmt.Printf("分配的内网 IP: 10.0.0.%d\n", i+2)
			fmt.Printf("客户端 Public Key: %s\n", peer.PublicKey)
			fmt.Printf("===============================\n")
			found = true
		}
	}

	if !found {
		fmt.Println("没有找到任何连接配置。")
	}
}
