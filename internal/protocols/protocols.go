package protocols

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/sail-tunnel/sbctl/internal/config"
	"github.com/sail-tunnel/sbctl/internal/generator"
)

func AddShadowsocks(cfg *config.SingBoxConfig, port int, remark string) (string, string, string, error) {
	serverPass, _ := generator.GeneratePassword()
	userPass, _ := generator.GeneratePassword()

	inbound := map[string]interface{}{
		"type":        "shadowsocks",
		"listen":      "::",
		"listen_port": port,
		"method":      "2022-blake3-aes-256-gcm",
		"password":    serverPass,
		"users": []interface{}{
			map[string]interface{}{
				"name":     remark,
				"password": userPass,
			},
		},
		"multiplex": map[string]interface{}{
			"enabled": true,
			"padding": true,
		},
	}
	cfg.Inbounds = append(cfg.Inbounds, inbound)
	return serverPass, userPass, "", nil
}

func AddVMess(cfg *config.SingBoxConfig, port int, remark string) (string, error) {
	uuid := generator.GenerateUUID()
	inbound := map[string]interface{}{
		"type":        "vmess",
		"listen":      "::",
		"listen_port": port,
		"users": []interface{}{
			map[string]interface{}{
				"name":    remark,
				"uuid":    uuid,
				"alterId": 0,
			},
		},
		"transport": map[string]interface{}{
			"type": "ws",
			"path": "/vmess",
		},
	}
	cfg.Inbounds = append(cfg.Inbounds, inbound)
	return uuid, nil
}

func AddHysteria2(cfg *config.SingBoxConfig, port int, remark string) (string, error) {
	password, _ := generator.GeneratePassword()
	certPath := config.DefaultConfigDir + "/certs/server.crt"
	keyPath := config.DefaultConfigDir + "/certs/server.key"
	if err := config.GenerateSelfSignedCert(certPath, keyPath); err != nil {
		return "", err
	}

	inbound := map[string]interface{}{
		"type":        "hysteria2",
		"listen":      "::",
		"listen_port": port,
		"users": []interface{}{
			map[string]interface{}{
				"name":     remark,
				"password": password,
			},
		},
		"tls": map[string]interface{}{
			"enabled":          true,
			"certificate_path": certPath,
			"key_path":         keyPath,
		},
	}
	cfg.Inbounds = append(cfg.Inbounds, inbound)
	return password, nil
}

func AddWireGuard(cfg *config.SingBoxConfig, port int, remark string) (string, string, string, string, error) {
	sPriv, sPub, _ := generator.GenerateWGKeyPair()
	cPriv, cPub, _ := generator.GenerateWGKeyPair()
	psk, _ := generator.GeneratePassword()

	ep := map[string]interface{}{
		"type":        "wireguard",
		"tag":         fmt.Sprintf("wg-in-%d", port),
		"listen_port": port,
		"system":      false,
		"name":        "wg0",
		"mtu":         1408,
		"address":     []string{"10.0.0.1/24", "fd00::1/64"},
		"private_key": sPriv,
		"peers": []interface{}{
			map[string]interface{}{
				"public_key":     cPub,
				"pre_shared_key": psk,
				"allowed_ips":    []string{"10.0.0.2/32", "fd00::2/128"},
				"comment":        remark,
			},
		},
	}
	cfg.Endpoints = append(cfg.Endpoints, ep)
	config.EnsureDirectInbound(cfg, port+1)
	return cPriv, cPub, sPub, psk, nil
}

// 打印分享链接
func PrintSS(ip string, port int, remark, serverPass, userPass string) {
	creds := fmt.Sprintf("2022-blake3-aes-256-gcm:%s:%s", serverPass, userPass)
	encoded := base64.StdEncoding.EncodeToString([]byte(creds))
	link := fmt.Sprintf("ss://%s@%s:%d#%s", encoded, ip, port, url.QueryEscape(remark))
	sub := base64.StdEncoding.EncodeToString([]byte(link))

	fmt.Printf("\n========== Shadowsocks 2022 ==========\n")
	fmt.Printf("服务器: %s  端口: %d  备注: %s\n", ip, port, remark)
	fmt.Printf("分享链接: %s\n", link)
	fmt.Printf("Base64 订阅: %s\n", sub)
	fmt.Printf("======================================\n")
}

func PrintVMess(ip string, port int, remark, uuid string) {
	jsonStr := fmt.Sprintf(`{"v":"2","ps":"%s","add":"%s","port":%d,"id":"%s","aid":0,"net":"ws","type":"none","host":"","path":"/vmess","tls":""}`, remark, ip, port, uuid)
	encoded := base64.StdEncoding.EncodeToString([]byte(jsonStr))
	link := fmt.Sprintf("vmess://%s", encoded)
	sub := base64.StdEncoding.EncodeToString([]byte(link))

	fmt.Printf("\n========== VMess + WebSocket ==========\n")
	fmt.Printf("服务器: %s  端口: %d  备注: %s\n", ip, port, remark)
	fmt.Printf("UUID: %s  传输: WebSocket /vmess\n", uuid)
	fmt.Printf("分享链接: %s\n", link)
	fmt.Printf("Base64 订阅: %s\n", sub)
	fmt.Printf("=======================================\n")
}

func PrintHysteria2(ip string, port int, remark, password string) {
	link := fmt.Sprintf("hy2://%s@%s:%d?insecure=1#%s", password, ip, port, url.QueryEscape(remark))
	sub := base64.StdEncoding.EncodeToString([]byte(link))

	fmt.Printf("\n========== Hysteria2 ==========\n")
	fmt.Printf("服务器: %s  端口: %d  备注: %s\n", ip, port, remark)
	fmt.Printf("注意: 使用自签名证书，客户端需开启 insecure/跳过验证\n")
	fmt.Printf("分享链接: %s\n", link)
	fmt.Printf("Base64 订阅: %s\n", sub)
	fmt.Printf("===============================\n")
}

func PrintWireGuard(ip string, port int, remark, clientPriv, serverPub, psk, userIP string) {
	fmt.Printf("\n========== WireGuard ==========\n")
	fmt.Printf("服务器: %s  端口: %d  备注: %s\n", ip, port, remark)
	fmt.Printf("客户端配置 (%s.conf):\n", remark)
	fmt.Printf("[Interface]\nPrivateKey = %s\nAddress = %s/24\nDNS = 8.8.8.8\n\n[Peer]\nPublicKey = %s\nPreSharedKey = %s\nEndpoint = %s:%d\nAllowedIPs = 0.0.0.0/0, ::/0\nPersistentKeepalive = 25\n",
		clientPriv, userIP, serverPub, psk, ip, port)
	fmt.Printf("===============================\n")
}
