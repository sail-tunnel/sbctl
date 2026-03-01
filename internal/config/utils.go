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

	"github.com/sagernet/sing-box/option"
)

// AppendInboundUser 给对应的端口（Inbound）追加用户
func AppendInboundUser(cfg *SingBoxConfig, port int, user interface{}) error {
	for i := range cfg.Inbounds {
		var listenPort uint16
		
		// 获取 listen_port
		switch opts := cfg.Inbounds[i].Options.(type) {
		case *option.ShadowsocksInboundOptions:
			listenPort = opts.ListenPort
		case *option.VMessInboundOptions:
			listenPort = opts.ListenPort
		case *option.Hysteria2InboundOptions:
			listenPort = opts.ListenPort
		default:
			continue
		}
		
		if listenPort == uint16(port) {
			switch cfg.Inbounds[i].Type {
			case "shadowsocks":
				if ssUser, ok := user.(option.ShadowsocksUser); ok {
					opts := cfg.Inbounds[i].Options.(*option.ShadowsocksInboundOptions)
					opts.Users = append(opts.Users, ssUser)
					return nil
				}
			case "vmess":
				if vmUser, ok := user.(option.VMessUser); ok {
					opts := cfg.Inbounds[i].Options.(*option.VMessInboundOptions)
					opts.Users = append(opts.Users, vmUser)
					return nil
				}
			case "hysteria2":
				if hy2User, ok := user.(option.Hysteria2User); ok {
					opts := cfg.Inbounds[i].Options.(*option.Hysteria2InboundOptions)
					opts.Users = append(opts.Users, hy2User)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("未在 config.json 的 inbounds 中找到监听端口为 %d 的配置", port)
}

// AppendEndpointPeer 给对应的端口（Endpoint/WireGuard）追加 Peer
func AppendEndpointPeer(cfg *SingBoxConfig, port int, peer option.WireGuardPeer) error {
	for i := range cfg.Endpoints {
		if cfg.Endpoints[i].Type == "wireguard" {
			if wgOpts, ok := cfg.Endpoints[i].Options.(*option.WireGuardEndpointOptions); ok {
				if wgOpts.ListenPort == uint16(port) {
					wgOpts.Peers = append(wgOpts.Peers, peer)
					return nil
				}
			}
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
