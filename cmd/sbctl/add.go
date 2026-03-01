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
// add 命令 - 向现有配置添加新协议
// =========================================================
func cmdAdd() {
	if !system.CheckRoot() {
		fmt.Println("[ERROR] 请使用 root 权限运行")
		os.Exit(1)
	}

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	var remark, protocol string
	var port int

	addCmd.StringVar(&remark, "r", "", "备注名")
	addCmd.StringVar(&remark, "remark", "", "备注名")
	addCmd.IntVar(&port, "p", 0, "端口")
	addCmd.IntVar(&port, "port", 0, "端口")
	addCmd.StringVar(&protocol, "P", "", "协议")
	addCmd.StringVar(&protocol, "protocol", "", "协议")

	addCmd.Parse(os.Args[2:])

	// 读取现有配置
	cfg, err := config.ReadConfig(config.DefaultConfigPath)
	if err != nil {
		fmt.Printf("[ERROR] 无法读取配置文件: %v\n", err)
		fmt.Println("[提示] 请先使用 'sbctl init' 初始化配置")
		os.Exit(1)
	}

	// 确保配置结构完整
	cfg, err = config.InitBaseConfig(config.DefaultConfigPath)
	if err != nil {
		fmt.Printf("[ERROR] 配置文件格式有误: %v\n", err)
		os.Exit(1)
	}

	if remark == "" {
		fmt.Print("请输入节点备注 (默认: myserver): ")
		fmt.Scanln(&remark)
		if remark == "" {
			remark = "myserver"
		}
	}

	if protocol == "" {
		fmt.Println("\n请选择协议:")
		fmt.Println("  1) Shadowsocks 2022  (推荐)")
		fmt.Println("  2) VMess + WebSocket")
		fmt.Println("  3) Hysteria2")
		fmt.Println("  4) WireGuard")
		fmt.Print("\n请输入序号 [1-4，默认 1]: ")
		var choice string
		fmt.Scanln(&choice)
		switch choice {
		case "2":
			protocol = "vmess"
		case "3":
			protocol = "hysteria2"
		case "4":
			protocol = "wireguard"
		default:
			protocol = "shadowsocks"
		}
	}

	if port == 0 {
		rdm, _ := generator.GenerateRandomPort()
		fmt.Printf("随机生成的端口为: %d , 回车确认或输入自定义端口: ", rdm)
		var custom string
		fmt.Scanln(&custom)
		if custom != "" {
			fmt.Sscanf(custom, "%d", &port)
		} else {
			port = rdm
		}
	}

	if port < 1 || port > 65535 {
		fmt.Println("[ERROR] 端口无效:", port)
		os.Exit(1)
	}

	// 检查端口冲突
	if config.CheckPortConflict(cfg, port) {
		fmt.Printf("[ERROR] 端口 %d 已经被占用！如果你想在此端口增加用户，请使用 'sbctl add-user -p %d'\n", port, port)
		os.Exit(1)
	}

	fmt.Printf("\n[INFO] 准备添加 -> 协议: %s  |  端口: %d  |  备注: %s\n", protocol, port, remark)
	ip := generator.GetPublicIP()

	switch protocol {
	case "shadowsocks":
		sp, up, _, _ := protocols.AddShadowsocks(cfg, port, remark)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintSS(ip, port, remark, sp, up)
	case "vmess":
		uuid, _ := protocols.AddVMess(cfg, port, remark)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintVMess(ip, port, remark, uuid)
	case "hysteria2":
		pw, _ := protocols.AddHysteria2(cfg, port, remark)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintHysteria2(ip, port, remark, pw)
	case "wireguard":
		cPriv, _, sPub, psk, _ := protocols.AddWireGuard(cfg, port, remark)
		config.WriteConfig(config.DefaultConfigPath, cfg)
		protocols.PrintWireGuard(ip, port, remark, cPriv, sPub, psk, "10.0.0.2")
	default:
		fmt.Println("[ERROR] 不支持的协议:", protocol)
		os.Exit(1)
	}

	system.RestartService()
	fmt.Println("[INFO] 配置已添加！已自动重启服务生效。")
}
