package main

import (
	"flag"
	"fmt"
	"os"

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

	cfg, err := config.ReadConfig(config.DefaultConfigPath)
	if err != nil {
		fmt.Println("[ERROR] 配置文件不存在或解析失败。请先运行 'sbctl init' 进行安装。")
		os.Exit(1)
	}

	ip := generator.GetPublicIP()
	found := false

	fmt.Println("\n=================================================")
	fmt.Printf(" 当前服务端 IP : %s\n", ip)
	fmt.Println("=================================================")

	// 遍历 inbounds
	for _, ib := range cfg.Inbounds {
		portFloat, ok := ib["listen_port"].(float64)
		if !ok {
			continue
		}
		port := int(portFloat)

		if targetPort != 0 && targetPort != port {
			continue
		}

		proto, _ := ib["type"].(string)

		users, ok := ib["users"].([]interface{})
		if !ok {
			continue
		}

		for _, u := range users {
			userMap := u.(map[string]interface{})
			remark, _ := userMap["name"].(string)

			switch proto {
			case "shadowsocks":
				serverPass, _ := ib["password"].(string)
				userPass, _ := userMap["password"].(string)
				protocols.PrintSS(ip, port, remark, serverPass, userPass)
				found = true

			case "vmess":
				uuid, _ := userMap["uuid"].(string)
				protocols.PrintVMess(ip, port, remark, uuid)
				found = true

			case "hysteria2":
				pw, _ := userMap["password"].(string)
				protocols.PrintHysteria2(ip, port, remark, pw)
				found = true
			}
		}
	}

	// 遍历 endpoints
	for _, ep := range cfg.Endpoints {
		portFloat, ok := ep["listen_port"].(float64)
		if !ok {
			continue
		}
		port := int(portFloat)

		if targetPort != 0 && targetPort != port {
			continue
		}

		proto, _ := ep["type"].(string)
		if proto == "wireguard" {
			peers, ok := ep["peers"].([]interface{})
			if !ok {
				continue
			}

			for i, p := range peers {
				peerMap := p.(map[string]interface{})
				remark, _ := peerMap["comment"].(string)
				if remark == "" {
					remark = fmt.Sprintf("Peer-%d", i)
				}

				fmt.Printf("\n========== WireGuard ==========\n")
				fmt.Printf("服务器: %s  端口: %d  备注: %s\n", ip, port, remark)
				fmt.Println("[WARN] 由于 WireGuard 私钥仅在生成时可知，无法重新完整展示客户端配置文件。")
				fmt.Printf("分配的内网 IP: 10.0.0.%d\n", i+2)
				if pub, ok := peerMap["public_key"].(string); ok {
					fmt.Printf("客户端 Public Key: %s\n", pub)
				}
				fmt.Printf("===============================\n")
				found = true
			}
		}
	}

	if !found {
		fmt.Println("没有找到任何连接配置。")
	}
}
