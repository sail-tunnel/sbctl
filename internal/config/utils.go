package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// AppendInboundUser 给对应的端口（Inbound）追加用户
func AppendInboundUser(cfg *SingBoxConfig, port int, user map[string]interface{}) error {
	for i, ib := range cfg.Inbounds {
		if p, ok := ib["listen_port"].(float64); ok && int(p) == port {
			var users []interface{}
			if u, ok := ib["users"].([]interface{}); ok {
				users = u
			}
			users = append(users, user)
			cfg.Inbounds[i]["users"] = users
			return nil
		}
	}
	return fmt.Errorf("未在 config.json 的 inbounds 中找到监听端口为 %d 的配置", port)
}

// AppendEndpointPeer 给对应的端口（Endpoint/WireGuard）追加 Peer
func AppendEndpointPeer(cfg *SingBoxConfig, port int, peer map[string]interface{}) error {
	for i, ep := range cfg.Endpoints {
		if p, ok := ep["listen_port"].(float64); ok && int(p) == port {
			var peers []interface{}
			if p, ok := ep["peers"].([]interface{}); ok {
				peers = p
			}
			peers = append(peers, peer)
			cfg.Endpoints[i]["peers"] = peers
			return nil
		}
	}
	return fmt.Errorf("未在 config.json 的 endpoints 中找到监听端口为 %d 的配置", port)
}

// GenerateSelfSignedCert 生成 Hysteria2 所需的自签名证书
func GenerateSelfSignedCert(certPath, keyPath string) error {
	if _, err := os.Stat(certPath); err == nil {
		return nil // 已存在
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "hysteria2",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 3650), // 10 years
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return nil
}
