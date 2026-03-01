package generator

import (
	"encoding/base64"
	"regexp"
	"testing"
)

// TestGeneratePassword 测试密码生成
func TestGeneratePassword(t *testing.T) {
	// 生成多个密码
	passwords := make(map[string]bool)
	for i := 0; i < 10; i++ {
		password, err := GeneratePassword()
		if err != nil {
			t.Fatalf("GeneratePassword failed: %v", err)
		}

		// 验证密码不为空
		if password == "" {
			t.Error("Generated password is empty")
		}

		// 验证是有效的 base64
		_, err = base64.StdEncoding.DecodeString(password)
		if err != nil {
			t.Errorf("Generated password is not valid base64: %v", err)
		}

		// 验证密码长度合理（32字节编码为base64约44字符）
		if len(password) < 40 {
			t.Errorf("Password too short: %d characters", len(password))
		}

		// 验证密码唯一性
		if passwords[password] {
			t.Error("Generated duplicate password")
		}
		passwords[password] = true
	}
}

// TestGenerateUUID 测试 UUID 生成
func TestGenerateUUID(t *testing.T) {
	// UUID 正则表达式（小写）
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// 生成多个 UUID
	uuids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		uuid := GenerateUUID()

		// 验证 UUID 不为空
		if uuid == "" {
			t.Error("Generated UUID is empty")
		}

		// 验证 UUID 格式
		if !uuidRegex.MatchString(uuid) {
			t.Errorf("Invalid UUID format: %s", uuid)
		}

		// 验证 UUID 唯一性
		if uuids[uuid] {
			t.Error("Generated duplicate UUID")
		}
		uuids[uuid] = true
	}
}

// TestGenerateRandomPort 测试随机端口生成
func TestGenerateRandomPort(t *testing.T) {
	// 生成多个端口
	ports := make(map[int]bool)
	for i := 0; i < 100; i++ {
		port, err := GenerateRandomPort()
		if err != nil {
			t.Fatalf("GenerateRandomPort failed: %v", err)
		}

		// 验证端口范围 10000-65535
		if port < 10000 || port > 65535 {
			t.Errorf("Port out of range: %d", port)
		}

		ports[port] = true
	}

	// 验证随机性（100次至少应该有80个不同的端口）
	if len(ports) < 80 {
		t.Errorf("Generated ports lack randomness: only %d unique ports out of 100", len(ports))
	}
}

// TestGetPublicIP 测试获取公网IP
func TestGetPublicIP(t *testing.T) {
	ip := GetPublicIP()

	// 验证返回值不为空
	if ip == "" {
		t.Error("GetPublicIP returned empty string")
	}

	// 验证返回的是合法 IP（简单验证）
	// 注意：在测试环境可能返回 "127.0.0.1"
	ipRegex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		t.Errorf("Invalid IP format: %s", ip)
	}

	t.Logf("Got public IP: %s", ip)
}

// TestGenerateWGKeyPair 测试 WireGuard 密钥对生成
func TestGenerateWGKeyPair(t *testing.T) {
	// 生成密钥对
	privateKey, publicKey, err := GenerateWGKeyPair()
	if err != nil {
		t.Fatalf("GenerateWGKeyPair failed: %v", err)
	}

	// 验证私钥不为空
	if privateKey == "" {
		t.Error("Generated private key is empty")
	}

	// 验证公钥不为空
	if publicKey == "" {
		t.Error("Generated public key is empty")
	}

	// 验证私钥是有效的 base64
	privBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		t.Errorf("Private key is not valid base64: %v", err)
	}

	// 验证私钥长度为 32 字节
	if len(privBytes) != 32 {
		t.Errorf("Private key wrong length: expected 32, got %d", len(privBytes))
	}

	// 验证公钥是有效的 base64
	pubBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		t.Errorf("Public key is not valid base64: %v", err)
	}

	// 验证公钥长度为 32 字节
	if len(pubBytes) != 32 {
		t.Errorf("Public key wrong length: expected 32, got %d", len(pubBytes))
	}

	// 验证唯一性：生成多个密钥对应该不同
	privateKey2, publicKey2, err := GenerateWGKeyPair()
	if err != nil {
		t.Fatalf("GenerateWGKeyPair (second) failed: %v", err)
	}

	if privateKey == privateKey2 {
		t.Error("Generated duplicate private key")
	}
	if publicKey == publicKey2 {
		t.Error("Generated duplicate public key")
	}
}

// TestGetWGPublicKey 测试从私钥派生公钥
func TestGetWGPublicKey(t *testing.T) {
	// 生成密钥对
	privateKey, expectedPublicKey, err := GenerateWGKeyPair()
	if err != nil {
		t.Fatalf("GenerateWGKeyPair failed: %v", err)
	}

	// 解码私钥
	privBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	var priv [32]byte
	copy(priv[:], privBytes)

	// 从私钥派生公钥
	pub := GetWGPublicKey(priv)
	derivedPublicKey := base64.StdEncoding.EncodeToString(pub[:])

	// 验证派生的公钥与原始公钥一致
	if derivedPublicKey != expectedPublicKey {
		t.Errorf("Derived public key mismatch:\nExpected: %s\nGot: %s", expectedPublicKey, derivedPublicKey)
	}
}

// TestGetWGPublicKeyFromString 测试从字符串私钥获取公钥
func TestGetWGPublicKeyFromString(t *testing.T) {
	// 生成密钥对
	privateKey, expectedPublicKey, err := GenerateWGKeyPair()
	if err != nil {
		t.Fatalf("GenerateWGKeyPair failed: %v", err)
	}

	// 从字符串私钥获取公钥
	publicKey, err := GetWGPublicKeyFromString(privateKey)
	if err != nil {
		t.Fatalf("GetWGPublicKeyFromString failed: %v", err)
	}

	// 验证公钥一致
	if publicKey != expectedPublicKey {
		t.Errorf("Public key mismatch:\nExpected: %s\nGot: %s", expectedPublicKey, publicKey)
	}
}

// TestGetWGPublicKeyFromString_InvalidInput 测试无效输入
func TestGetWGPublicKeyFromString_InvalidInput(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"invalid base64", "not-base64!!!"},
		{"wrong length", base64.StdEncoding.EncodeToString([]byte("short"))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetWGPublicKeyFromString(tc.input)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}
		})
	}
}

// BenchmarkGeneratePassword 性能测试：密码生成
func BenchmarkGeneratePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GeneratePassword()
	}
}

// BenchmarkGenerateUUID 性能测试：UUID 生成
func BenchmarkGenerateUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateUUID()
	}
}

// BenchmarkGenerateRandomPort 性能测试：随机端口生成
func BenchmarkGenerateRandomPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRandomPort()
	}
}

// BenchmarkGenerateWGKeyPair 性能测试：WireGuard 密钥对生成
func BenchmarkGenerateWGKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = GenerateWGKeyPair()
	}
}
