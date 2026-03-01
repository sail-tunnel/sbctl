package protocols

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"net/url"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sail-tunnel/sbctl/internal/config"
	"github.com/sail-tunnel/sbctl/internal/generator"
)

func AddShadowsocks(cfg *config.SingBoxConfig, port int, remark string) (string, string, string, error) {
	serverPass, _ := generator.GeneratePassword()
	userPass, _ := generator.GeneratePassword()

	listenAddr, _ := netip.ParseAddr("::")
	badAddr := (*badoption.Addr)(&listenAddr)

	inbound := option.Inbound{
		Type: "shadowsocks",
		Tag:  fmt.Sprintf("ss-in-%d", port),
		Options: &option.ShadowsocksInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: uint16(port),
			},
			Method:   "2022-blake3-aes-256-gcm",
			Password: serverPass,
			Users: []option.ShadowsocksUser{
				{
					Name:     remark,
					Password: userPass,
				},
			},
			Multiplex: &option.InboundMultiplexOptions{
				Enabled: true,
				Padding: true,
			},
		},
	}
	cfg.Inbounds = append(cfg.Inbounds, inbound)
	return serverPass, userPass, "", nil
}

func AddVMess(cfg *config.SingBoxConfig, port int, remark string) (string, error) {
	uuid := generator.GenerateUUID()
	listenAddr, _ := netip.ParseAddr("::")
	badAddr := (*badoption.Addr)(&listenAddr)

	inbound := option.Inbound{
		Type: "vmess",
		Tag:  fmt.Sprintf("vmess-in-%d", port),
		Options: &option.VMessInboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: uint16(port),
			},
			Users: []option.VMessUser{
				{
					Name: remark,
					UUID: uuid,
				},
			},
			Transport: &option.V2RayTransportOptions{
				Type: "ws",
				WebsocketOptions: option.V2RayWebsocketOptions{
					Path: "/vmess",
				},
			},
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

	listenAddr, _ := netip.ParseAddr("::")
	badAddr := (*badoption.Addr)(&listenAddr)

	inbound := option.Inbound{
		Type: "hysteria2",
		Tag:  fmt.Sprintf("hy2-in-%d", port),
		Options: &option.Hysteria2InboundOptions{
			ListenOptions: option.ListenOptions{
				Listen:     badAddr,
				ListenPort: uint16(port),
			},
			Users: []option.Hysteria2User{
				{
					Name:     remark,
					Password: password,
				},
			},
			InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{
				TLS: &option.InboundTLSOptions{
					Enabled:         true,
					CertificatePath: certPath,
					KeyPath:         keyPath,
				},
			},
		},
	}
	cfg.Inbounds = append(cfg.Inbounds, inbound)
	return password, nil
}

func AddWireGuard(cfg *config.SingBoxConfig, port int, remark string) (string, string, string, string, error) {
	sPriv, sPub, _ := generator.GenerateWGKeyPair()
	cPriv, cPub, _ := generator.GenerateWGKeyPair()
	psk, _ := generator.GeneratePassword()

	addr1, _ := netip.ParsePrefix("10.0.0.1/24")
	addr2, _ := netip.ParsePrefix("fd00::1/64")
	peerAddr1, _ := netip.ParsePrefix("10.0.0.2/32")
	peerAddr2, _ := netip.ParsePrefix("fd00::2/128")

	endpoint := option.Endpoint{
		Type: "wireguard",
		Tag:  fmt.Sprintf("wg-in-%d", port),
		Options: &option.WireGuardEndpointOptions{
			System:     false,
			Name:       "wg0",
			MTU:        1408,
			Address:    []netip.Prefix{addr1, addr2},
			PrivateKey: sPriv,
			ListenPort: uint16(port),
			Peers: []option.WireGuardPeer{
				{
					PublicKey:    cPub,
					PreSharedKey: psk,
					AllowedIPs:   []netip.Prefix{peerAddr1, peerAddr2},
				},
			},
		},
	}
	cfg.Endpoints = append(cfg.Endpoints, endpoint)
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
