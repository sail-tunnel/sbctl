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

func cmdUpgrade() {
	if !system.CheckRoot() {
		fmt.Println("[ERROR] 请使用 root 权限运行升级: sudo sbctl upgrade")
		os.Exit(1)
	}

	arch := runtime.GOARCH
	// 映射内核架构到下载名
	dlArch := arch
	if arch == "amd64" {
		dlArch = "amd64"
	} else if arch == "arm64" {
		dlArch = "arm64"
	} else if arch == "arm" {
		dlArch = "arm"
	} else {
		fmt.Printf("[ERROR] 不支持的架构: %s\n", arch)
		os.Exit(1)
	}

	downloadURL := RepoURL + dlArch
	fmt.Printf("[INFO] 正在从 GitHub 下载最新版本 (%s)...\n", dlArch)

	resp, err := http.Get(downloadURL)
	if err != nil {
		fmt.Printf("[ERROR] 下载失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[ERROR] 下载失败，服务器返回状态码: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// 获取当前运行的可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("[ERROR] 无法获取当前程序路径: %v\n", err)
		os.Exit(1)
	}

	// 先下载到临时文件
	tmpFile := exePath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		fmt.Printf("[ERROR] 无法创建临时文件: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("[ERROR] 写入临时文件失败: %v\n", err)
		os.Exit(1)
	}

	// 给临时文件可执行权限
	if err := os.Chmod(tmpFile, 0755); err != nil {
		fmt.Printf("[ERROR] 无法设置权限: %v\n", err)
		os.Exit(1)
	}

	// 替换原始文件 (Linux 下运行中的二进制文件可以直接被 rename 替换)
	if err := os.Rename(tmpFile, exePath); err != nil {
		fmt.Printf("[ERROR] 替换程序失败: %v\n", err)
		// 尝试通过 mv 命令强行替换
		cmd := exec.Command("mv", "-f", tmpFile, exePath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("[ERROR] 尝试使用 mv 替换也失败了: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("[SUCCESS] sbctl 升级成功！")

	// 输出当前版本信息 (如果有版本号记录的话)
	// fmt.Printf("当前版本: %s\n", Version)
}
