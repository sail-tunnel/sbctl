package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/sail-tunnel/sbctl/internal/system"
)

const (
	RepoURL = "https://github.com/sail-tunnel/sbctl/releases/latest/download/sbctl-linux-"
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorReset  = "\033[0m"
)

func logInfo(msg string)  { fmt.Printf("%s[INFO]%s  %s\n", colorGreen, colorReset, msg) }
func logWarn(msg string)  { fmt.Printf("%s[WARN]%s  %s\n", colorYellow, colorReset, msg) }
func logError(msg string) { fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, msg) }

func cmdUpgrade() {
	if !system.CheckRoot() {
		logError("请使用 root 权限运行: sudo sbctl upgrade")
		os.Exit(1)
	}

	switch runtime.GOARCH {
	case "amd64", "arm64", "arm":
	default:
		logError("不支持的架构: " + runtime.GOARCH)
		os.Exit(1)
	}
	dlArch := runtime.GOARCH

	fmt.Println("=== sbctl 升级 ===")
	fmt.Println()
	logInfo("当前版本: " + version)

	downloadURL := RepoURL + dlArch
	logInfo(fmt.Sprintf("正在下载最新版本 (%s)...", dlArch))

	resp, err := http.Get(downloadURL)
	if err != nil {
		logError("下载失败: " + err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logError(fmt.Sprintf("下载失败，服务器返回状态码: %d", resp.StatusCode))
		os.Exit(1)
	}

	exePath, err := os.Executable()
	if err != nil {
		logError("无法获取当前程序路径: " + err.Error())
		os.Exit(1)
	}

	tmpFile := exePath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		logError("无法创建临时文件: " + err.Error())
		os.Exit(1)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		logError("写入临时文件失败: " + err.Error())
		os.Exit(1)
	}
	out.Close()

	if err := os.Chmod(tmpFile, 0755); err != nil {
		logError("无法设置权限: " + err.Error())
		os.Exit(1)
	}

	if err := os.Rename(tmpFile, exePath); err != nil {
		cmd := exec.Command("mv", "-f", tmpFile, exePath)
		if err := cmd.Run(); err != nil {
			logError("替换程序失败: " + err.Error())
			os.Exit(1)
		}
	}

	fmt.Println()
	logInfo(fmt.Sprintf("%s升级完成！%s 使用 'sbctl help' 查看所有命令", colorGreen, colorReset))
}
