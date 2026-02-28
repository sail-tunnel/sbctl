package system

import (
	"os"
	"os/exec"
)

// CheckRoot 验证当前是否为 Root
func CheckRoot() bool {
	return os.Geteuid() == 0
}

// RestartService 尝试重启 sing-box
func RestartService() error {
	_ = exec.Command("systemctl", "enable", "sing-box").Run()
	return exec.Command("systemctl", "restart", "sing-box").Run()
}

func StartService() error {
	return exec.Command("systemctl", "start", "sing-box").Run()
}

func StopService() error {
	return exec.Command("systemctl", "stop", "sing-box").Run()
}

func ViewLogs() error {
	cmd := exec.Command("journalctl", "-u", "sing-box", "-n", "100", "--no-pager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ViewStatus() error {
	cmd := exec.Command("systemctl", "status", "sing-box", "--no-pager")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
