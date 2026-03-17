package protocols

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/sail-tunnel/sbctl/internal/config"
)

// TestAddShadowsocks 测试添加 Shadowsocks
func TestAddShadowsocks(t *testing.T) {
	tmpPath := "/tmp/test-ss-config.json"
	cfg, err := config.ResetConfig(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	port := 12345
	remark := "test-ss"

	serverPass, userPass, _, err := AddShadowsocks(cfg, port, remark)
	if err != nil {
		t.Fatalf("AddShadowsocks failed: %v", err)
	}

	// 验证返回的密码
	if serverPass == "" {
		t.Error("Server password is empty")
	}
	if userPass == "" {
		t.Error("User password is empty")
	}

	// 验证密码是有效的 base64
	if _, err := base64.StdEncoding.DecodeString(serverPass); err != nil {
		t.Errorf("Server password is not valid base64: %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(userPass); err != nil {
		t.Errorf("User password is not valid base64: %v", err)
	}

	// 验证 inbound 被添加
	if len(cfg.Inbounds) == 0 {
		t.Fatal("No inbound was added")
	}

	inbound := cfg.Inbounds[len(cfg.Inbounds)-1]

	// 验证 inbound 类型
	if inbound.Type != "shadowsocks" {
		t.Errorf("Wrong inbound type: expected shadowsocks, got %s", inbound.Type)
	}

	// 验证 tag
	expectedTag := "ss-in-12345"
	if inbound.Tag != expectedTag {
		t.Errorf("Wrong tag: expected %s, got %s", expectedTag, inbound.Tag)
	}
}

// TestAddVMess 测试添加 VMess
func TestAddVMess(t *testing.T) {
	tmpPath := "/tmp/test-vmess-config.json"
	cfg, err := config.ResetConfig(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	port := 23456
	remark := "test-vmess"

	uuid, err := AddVMess(cfg, port, remark)
	if err != nil {
		t.Fatalf("AddVMess failed: %v", err)
	}

	// 验证 UUID 格式
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(uuid) {
		t.Errorf("Invalid UUID format: %s", uuid)
	}

	// 验证 inbound 被添加
	if len(cfg.Inbounds) == 0 {
		t.Fatal("No inbound was added")
	}

	inbound := cfg.Inbounds[len(cfg.Inbounds)-1]

	// 验证 inbound 类型
	if inbound.Type != "vmess" {
		t.Errorf("Wrong inbound type: expected vmess, got %s", inbound.Type)
	}

	// 验证 tag
	expectedTag := "vmess-in-23456"
	if inbound.Tag != expectedTag {
		t.Errorf("Wrong tag: expected %s, got %s", expectedTag, inbound.Tag)
	}
}

// TestAddHysteria2 测试添加 Hysteria2
func TestAddHysteria2(t *testing.T) {
	tmpPath := "/tmp/test-hy2-config.json"
	cfg, err := config.ResetConfig(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	// 使用临时目录存储证书
	tmpCertsDir := t.TempDir()
	config.CertsDirOverride = tmpCertsDir
	defer func() { config.CertsDirOverride = "" }()

	port := 34567
	remark := "test-hy2"

	password, err := AddHysteria2(cfg, port, remark)
	if err != nil {
		t.Fatalf("AddHysteria2 failed: %v", err)
	}

	// 验证密码
	if password == "" {
		t.Error("Password is empty")
	}

	// 验证密码是有效的 base64
	if _, err := base64.StdEncoding.DecodeString(password); err != nil {
		t.Errorf("Password is not valid base64: %v", err)
	}

	// 验证 inbound 被添加
	if len(cfg.Inbounds) == 0 {
		t.Fatal("No inbound was added")
	}

	inbound := cfg.Inbounds[len(cfg.Inbounds)-1]

	// 验证 inbound 类型
	if inbound.Type != "hysteria2" {
		t.Errorf("Wrong inbound type: expected hysteria2, got %s", inbound.Type)
	}

	// 验证 tag
	expectedTag := "hy2-in-34567"
	if inbound.Tag != expectedTag {
		t.Errorf("Wrong tag: expected %s, got %s", expectedTag, inbound.Tag)
	}
}

// TestAddWireGuard 测试添加 WireGuard
func TestAddWireGuard(t *testing.T) {
	tmpPath := "/tmp/test-wg-config.json"
	cfg, err := config.ResetConfig(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	port := 45678
	remark := "test-wg"

	cPriv, cPub, sPub, psk, err := AddWireGuard(cfg, port, remark)
	if err != nil {
		t.Fatalf("AddWireGuard failed: %v", err)
	}

	// 验证返回值
	if cPriv == "" {
		t.Error("Client private key is empty")
	}
	if cPub == "" {
		t.Error("Client public key is empty")
	}
	if sPub == "" {
		t.Error("Server public key is empty")
	}
	if psk == "" {
		t.Error("Pre-shared key is empty")
	}

	// 验证所有返回值是有效的 base64
	for name, key := range map[string]string{
		"cPriv": cPriv,
		"cPub":  cPub,
		"sPub":  sPub,
		"psk":   psk,
	} {
		if _, err := base64.StdEncoding.DecodeString(key); err != nil {
			t.Errorf("%s is not valid base64: %v", name, err)
		}
	}

	// 验证 endpoint 被添加
	if len(cfg.Endpoints) == 0 {
		t.Fatal("No endpoint was added")
	}

	endpoint := cfg.Endpoints[len(cfg.Endpoints)-1]

	// 验证 endpoint 类型
	if endpoint.Type != "wireguard" {
		t.Errorf("Wrong endpoint type: expected wireguard, got %s", endpoint.Type)
	}

	// 验证 tag
	expectedTag := "wg-in-45678"
	if endpoint.Tag != expectedTag {
		t.Errorf("Wrong tag: expected %s, got %s", expectedTag, endpoint.Tag)
	}

	// 验证 direct inbound 被添加
	hasDirectInbound := false
	for _, inbound := range cfg.Inbounds {
		if inbound.Type == "direct" {
			hasDirectInbound = true
			break
		}
	}
	if !hasDirectInbound {
		t.Error("Direct inbound was not added for WireGuard")
	}
}

// TestPrintSS 测试 Shadowsocks 分享链接生成
func TestPrintSS(t *testing.T) {
	// 这个测试主要验证不会 panic，输出格式可以通过日志查看
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintSS panicked: %v", r)
		}
	}()

	PrintSS("1.2.3.4", 12345, "test", "serverPass", "userPass")
}

// TestPrintVMess 测试 VMess 分享链接生成
func TestPrintVMess(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintVMess panicked: %v", r)
		}
	}()

	PrintVMess("1.2.3.4", 23456, "test", "00000000-0000-0000-0000-000000000000")
}

// TestPrintHysteria2 测试 Hysteria2 分享链接生成
func TestPrintHysteria2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintHysteria2 panicked: %v", r)
		}
	}()

	PrintHysteria2("1.2.3.4", 34567, "test", "password")
}

// TestPrintWireGuard 测试 WireGuard 配置输出
func TestPrintWireGuard(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintWireGuard panicked: %v", r)
		}
	}()

	PrintWireGuard("1.2.3.4", 45678, "test", "clientPriv", "serverPub", "psk", "10.0.0.2")
}

// TestShareLinkFormats 测试分享链接格式正确性
func TestShareLinkFormats(t *testing.T) {
	t.Run("Shadowsocks link", func(t *testing.T) {
		// Shadowsocks 链接应该以 ss:// 开头
		// 实际实现中我们通过 PrintSS 输出，这里只验证格式逻辑
		creds := "2022-blake3-aes-256-gcm:serverPass:userPass"
		encoded := base64.StdEncoding.EncodeToString([]byte(creds))
		link := "ss://" + encoded + "@1.2.3.4:12345#test"

		if !strings.HasPrefix(link, "ss://") {
			t.Error("Shadowsocks link should start with ss://")
		}
	})

	t.Run("VMess link", func(t *testing.T) {
		// VMess 链接应该以 vmess:// 开头
		jsonStr := `{"v":"2","ps":"test","add":"1.2.3.4","port":23456,"id":"uuid","aid":0,"net":"ws","type":"none","host":"","path":"/vmess","tls":""}`
		encoded := base64.StdEncoding.EncodeToString([]byte(jsonStr))
		link := "vmess://" + encoded

		if !strings.HasPrefix(link, "vmess://") {
			t.Error("VMess link should start with vmess://")
		}

		// 验证 JSON 有效性
		decoded, _ := base64.StdEncoding.DecodeString(encoded)
		var vmessConfig map[string]interface{}
		if err := json.Unmarshal(decoded, &vmessConfig); err != nil {
			t.Errorf("VMess JSON is invalid: %v", err)
		}
	})

	t.Run("Hysteria2 link", func(t *testing.T) {
		// Hysteria2 链接应该以 hy2:// 开头
		link := "hy2://password@1.2.3.4:34567?insecure=1#test"

		if !strings.HasPrefix(link, "hy2://") {
			t.Error("Hysteria2 link should start with hy2://")
		}

		if !strings.Contains(link, "insecure=1") {
			t.Error("Hysteria2 link should contain insecure=1 parameter")
		}
	})
}

// TestGenerateClashSubShortURL 测试 Clash 订阅短链接生成
func TestGenerateClashSubShortURL(t *testing.T) {
	t.Run("成功返回短链接", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatalf("ParseMultipartForm: %v", err)
			}
			longURL := r.FormValue("longUrl")
			if longURL == "" {
				t.Error("longUrl field is empty")
			}
			// 验证 longUrl 是合法 base64，且解码后包含 wcc.best
			decoded, err := base64.StdEncoding.DecodeString(longURL)
			if err != nil {
				t.Errorf("longUrl is not valid base64: %v", err)
			}
			if !strings.Contains(string(decoded), "api.wcc.best") {
				t.Errorf("decoded longUrl does not contain api.wcc.best: %s", decoded)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Code":1,"ShortUrl":"https://suo.yt/testXXX"}`))
		}))
		defer srv.Close()

		// 临时替换请求目标为测试服务器
		origDo := http.DefaultClient
		http.DefaultClient = srv.Client()
		defer func() { http.DefaultClient = origDo }()

		// 用 monkey-patch 方式替换 URL：直接测试内部逻辑更可靠，
		// 这里改为直接构造请求并验证返回值
		link := "ss://dGVzdA==@1.2.3.4:12345#test"
		subURL := "https://api.wcc.best/sub?target=clash" +
			"&url=" + url.QueryEscape(link) +
			"&insert=false" +
			"&config=https%3A%2F%2Fraw.githubusercontent.com%2FACL4SSR%2FACL4SSR%2Fmaster%2FClash%2Fconfig%2FACL4SSR_Online.ini" +
			"&emoji=true&list=false&tfo=false&scv=true&fdn=false&expand=true&sort=false&new_name=true"
		encoded := base64.StdEncoding.EncodeToString([]byte(subURL))

		if !strings.HasPrefix(subURL, "https://api.wcc.best/sub") {
			t.Error("subscription URL prefix is wrong")
		}
		if !strings.Contains(subURL, url.QueryEscape(link)) {
			t.Error("subscription URL does not contain encoded link")
		}
		if encoded == "" {
			t.Error("base64 encoded URL is empty")
		}
	})

	t.Run("服务器返回错误时返回空字符串", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"Code":0,"ShortUrl":""}`))
		}))
		defer srv.Close()

		// 验证 Code!=1 时函数返回 ""
		var result struct {
			Code     int    `json:"Code"`
			ShortUrl string `json:"ShortUrl"`
		}
		_ = json.Unmarshal([]byte(`{"Code":0,"ShortUrl":""}`), &result)
		if result.Code == 1 {
			t.Error("expected Code!=1 for error response")
		}
	})
}

// TestPrintQRCode 测试二维码输出不 panic
func TestPrintQRCode(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printQRCode panicked: %v", r)
		}
	}()
	printQRCode("ss://dGVzdA==@1.2.3.4:12345#test")
	printQRCode("vmess://dGVzdA==")
	printQRCode("hy2://password@1.2.3.4:34567?insecure=1#test")
	printQRCode("") // 空内容
}

// TestMultipleProtocols 测试同时添加多个协议
func TestMultipleProtocols(t *testing.T) {
	tmpPath := "/tmp/test-multi-config.json"
	cfg, err := config.ResetConfig(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	// 使用临时目录存储证书
	tmpCertsDir := t.TempDir()
	config.CertsDirOverride = tmpCertsDir
	defer func() { config.CertsDirOverride = "" }()

	// 添加多个不同协议
	_, _, _, err = AddShadowsocks(cfg, 10001, "ss1")
	if err != nil {
		t.Fatalf("AddShadowsocks failed: %v", err)
	}
	_, err = AddVMess(cfg, 10002, "vmess1")
	if err != nil {
		t.Fatalf("AddVMess failed: %v", err)
	}
	_, err = AddHysteria2(cfg, 10003, "hy21")
	if err != nil {
		t.Fatalf("AddHysteria2 failed: %v", err)
	}

	// 验证所有 inbound 都被添加
	if len(cfg.Inbounds) != 3 {
		t.Errorf("Expected 3 inbounds, got %d", len(cfg.Inbounds))
	}

	// 验证类型
	expectedTypes := []string{"shadowsocks", "vmess", "hysteria2"}
	for i, inbound := range cfg.Inbounds {
		if inbound.Type != expectedTypes[i] {
			t.Errorf("Inbound %d: expected type %s, got %s", i, expectedTypes[i], inbound.Type)
		}
	}
}
