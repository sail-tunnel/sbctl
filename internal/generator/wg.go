package generator

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// GenerateWGKeyPair 使用 Curve25519 直接生成 WireGuard 公私钥对 (无第三方依赖)
func GenerateWGKeyPair() (privateKey string, publicKey string, err error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return "", "", err
	}
	// 按照 X25519 规范裁剪
	priv[0] &= 248
	priv[31] = (priv[31] & 127) | 64

	pub := GetWGPublicKey(priv)

	privateKeyBase64 := base64.StdEncoding.EncodeToString(priv[:])
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pub[:])

	return privateKeyBase64, publicKeyBase64, nil
}

// GetWGPublicKey 从私钥派生公钥
func GetWGPublicKey(priv [32]byte) [32]byte {
	var pub [32]byte
	curve25519.ScalarBaseMult(&pub, &priv)
	return pub
}

// GetWGPublicKeyFromString 从 base64 私钥字符串获取公钥字符串
func GetWGPublicKeyFromString(privBase64 string) (string, error) {
	priv, err := base64.StdEncoding.DecodeString(privBase64)
	if err != nil || len(priv) != 32 {
		return "", fmt.Errorf("invalid private key")
	}
	var privBytes [32]byte
	copy(privBytes[:], priv)
	pub := GetWGPublicKey(privBytes)
	return base64.StdEncoding.EncodeToString(pub[:]), nil
}
