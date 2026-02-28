package main

import (
	"fmt"
	"os"

	"github.com/sail-tunnel/sbctl/internal/system"
)

func showHelp() {
	fmt.Printf(`sbctl v2.0.0 (Go Edition) - Sing-Box 多协议多用户管理工具

命令用法:
    sbctl init <选项>     初始化或追加新协议节点
    sbctl add-user <选项> 向已有协议端口追加新客户端(用户)
    sbctl show [选项]     显示当前配置和订阅链接
    sbctl status         查看服务端运行状态
    sbctl restart        重启 sing-box 服务
    sbctl upgrade        升级 sbctl 核心程序
    sbctl start          启动 sing-box 服务
    sbctl stop           停止 sing-box 服务
    sbctl log            查看服务运行日志

init 选项 (新节点):
    -r, --remark REMARK  自定义节点备注名称 (默认: myserver)
    -p, --port PORT      指定端口号 (如果不指定则随机生成)
    -P, --protocol PROTO 选择协议 (shadowsocks/vmess/hysteria2/wireguard)

add-user 选项 (新用户):
    -p, --port PORT      目标协议监听的端口
    -r, --remark REMARK  新用户的备注

show 选项:
    -p, --port PORT      仅展示特定端口下的节点信息
`)
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		cmdInit()
	case "add-user":
		cmdAddUser()
	case "show":
		cmdShow()
	case "status":
		system.ViewStatus()
	case "start":
		if !system.CheckRoot() {
			fmt.Println("[ERROR] 请使用 root 权限运行")
			os.Exit(1)
		}
		if err := system.StartService(); err == nil {
			fmt.Println("[INFO] 服务已启动")
		} else {
			fmt.Println("[ERROR] 启动失败:", err)
		}
	case "stop":
		if !system.CheckRoot() {
			fmt.Println("[ERROR] 请使用 root 权限运行")
			os.Exit(1)
		}
		if err := system.StopService(); err == nil {
			fmt.Println("[INFO] 服务已停止")
		} else {
			fmt.Println("[ERROR] 停止失败:", err)
		}
	case "restart":
		if !system.CheckRoot() {
			fmt.Println("[ERROR] 请使用 root 权限运行")
			os.Exit(1)
		}
		if err := system.RestartService(); err == nil {
			fmt.Println("[INFO] 服务已重启")
		} else {
			fmt.Println("[ERROR] 重启失败:", err)
		}
	case "upgrade":
		cmdUpgrade()
	case "log":
		system.ViewLogs()
	case "help", "--help", "-h":
		showHelp()
	default:
		fmt.Printf("[ERROR] 未知命令: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
}
