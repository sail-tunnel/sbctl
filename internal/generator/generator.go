package generator

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// GeneratePassword 生成32位随机 base64 密码
func GeneratePassword() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// GenerateUUID 生成小写标准 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomPort 生成 10000 到 65535 的随机端口
func GenerateRandomPort() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(55536))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 10000, nil
}

// GetPublicIP 获取当前公网 IP
func GetPublicIP() string {
	apis := []string{
		"http://ifconfig.me",
		"http://icanhazip.com",
		"http://api.ipify.org",
	}

	for _, api := range apis {
		resp, err := http.Get(api)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					return strings.TrimSpace(string(body))
				}
			}
		}
	}
	return "127.0.0.1" // Fallback
}
