package generator

import (
	"crypto/rand"
	"encoding/base64"

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

	var pub [32]byte
	curve25519.ScalarBaseMult(&pub, &priv)

	privateKeyBase64 := base64.StdEncoding.EncodeToString(priv[:])
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pub[:])

	return privateKeyBase64, publicKeyBase64, nil
}
