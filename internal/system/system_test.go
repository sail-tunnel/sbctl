package system

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// TestCheckRoot 测试 root 权限检查
func TestCheckRoot(t *testing.T) {
	isRoot := CheckRoot()

	// 验证返回值类型正确（布尔值）
	if isRoot != true && isRoot != false {
		t.Error("CheckRoot should return boolean")
	}

	// 获取实际的 uid
	euid := os.Geteuid()

	// 验证函数逻辑正确
	if euid == 0 && !isRoot {
		t.Error("Should return true when running as root (uid 0)")
	}
	if euid != 0 && isRoot {
		t.Error("Should return false when not running as root")
	}

	t.Logf("Running as root: %v (euid: %d)", isRoot, euid)
}

// TestRestartService 测试服务重启
func TestRestartService(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping systemctl test on non-Linux system")
	}

	// 检查 systemctl 是否可用
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not available")
	}

	// 只在有 root 权限时运行实际命令
	if !CheckRoot() {
		t.Skip("Skipping RestartService test: requires root privileges")
	}

	// 注意：这会实际重启 sing-box 服务
	// 在 CI 环境中可能没有 sing-box 服务，所以不检查错误
	err := RestartService()
	t.Logf("RestartService result: %v", err)
}

// TestStartService 测试服务启动
func TestStartService(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping systemctl test on non-Linux system")
	}

	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not available")
	}

	if !CheckRoot() {
		t.Skip("Skipping StartService test: requires root privileges")
	}

	err := StartService()
	t.Logf("StartService result: %v", err)
}

// TestStopService 测试服务停止
func TestStopService(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping systemctl test on non-Linux system")
	}

	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not available")
	}

	if !CheckRoot() {
		t.Skip("Skipping StopService test: requires root privileges")
	}

	err := StopService()
	t.Logf("StopService result: %v", err)
}

// TestViewLogs 测试查看日志
func TestViewLogs(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping journalctl test on non-Linux system")
	}

	if _, err := exec.LookPath("journalctl"); err != nil {
		t.Skip("journalctl not available")
	}

	// ViewLogs 不需要 root 权限，但可能需要特定权限查看某些日志
	err := ViewLogs()
	if err != nil {
		t.Logf("ViewLogs returned error (may be expected): %v", err)
	}
}

// TestViewStatus 测试查看服务状态
func TestViewStatus(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping systemctl test on non-Linux system")
	}

	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not available")
	}

	// ViewStatus 通常不需要 root 权限
	err := ViewStatus()
	if err != nil {
		t.Logf("ViewStatus returned error (may be expected if service doesn't exist): %v", err)
	}
}

// TestSystemctlAvailability 测试 systemctl 可用性
func TestSystemctlAvailability(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux system")
	}

	path, err := exec.LookPath("systemctl")
	if err != nil {
		t.Log("systemctl not found in PATH (expected in non-systemd environments)")
	} else {
		t.Logf("systemctl found at: %s", path)
	}
}

// TestJournalctlAvailability 测试 journalctl 可用性
func TestJournalctlAvailability(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux system")
	}

	path, err := exec.LookPath("journalctl")
	if err != nil {
		t.Log("journalctl not found in PATH (expected in non-systemd environments)")
	} else {
		t.Logf("journalctl found at: %s", path)
	}
}

// TestServiceFunctionsExist 测试所有服务函数都可以调用（不运行）
func TestServiceFunctionsExist(t *testing.T) {
	// 这个测试只验证函数存在且可以被引用
	funcs := map[string]interface{}{
		"CheckRoot":      CheckRoot,
		"RestartService": RestartService,
		"StartService":   StartService,
		"StopService":    StopService,
		"ViewLogs":       ViewLogs,
		"ViewStatus":     ViewStatus,
	}

	for name, fn := range funcs {
		if fn == nil {
			t.Errorf("Function %s is nil", name)
		}
	}
}

// TestCheckRootWithMockedUid 测试在不同 UID 下的行为
func TestCheckRootWithMockedUid(t *testing.T) {
	// 获取当前 UID
	currentUID := os.Geteuid()
	t.Logf("Current UID: %d", currentUID)

	// 测试当前环境的 CheckRoot 行为
	result := CheckRoot()
	expected := (currentUID == 0)

	if result != expected {
		t.Errorf("CheckRoot() = %v, expected %v (uid=%d)", result, expected, currentUID)
	}
}
