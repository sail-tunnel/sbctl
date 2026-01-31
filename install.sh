#!/bin/bash

# sbctl - Sing-Box 一键安装脚本
# https://github.com/sail-tunnel/sbctl

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 root 权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 权限运行: sudo $0"
        exit 1
    fi
}

# 生成随机密码
generate_password() {
    openssl rand -base64 32
}

# 获取公网 IP
get_public_ip() {
    local ip
    ip=$(curl -s -4 ifconfig.me) || ip=$(curl -s -4 icanhazip.com) || ip=$(curl -s -4 api.ipify.org)
    echo "$ip"
}

# 验证邮箱格式
validate_email() {
    if [[ ! "$1" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        return 1
    fi
    return 0
}

# 验证端口
validate_port() {
    if [[ ! "$1" =~ ^[0-9]+$ ]] || [ "$1" -lt 1 ] || [ "$1" -gt 65535 ]; then
        return 1
    fi
    return 0
}

# 安装基础工具
install_base_packages() {
    log_info "安装基础工具..."
    apt update
    apt install -y curl vim
}

# 安装 BBR v3
install_bbr() {
    log_info "安装 BBR v3（可选）..."
    if curl -L -s https://raw.githubusercontent.com/byJoey/Actions-bbr-v3/refs/heads/main/install.sh -o /tmp/bbr-install.sh; then
        echo "1" | bash /tmp/bbr-install.sh || log_warn "BBR v3 安装失败，跳过"
    else
        log_warn "BBR v3 脚本下载失败，跳过"
    fi
}

# 安装 sing-box
install_singbox() {
    log_info "安装 sing-box..."
    
    # 添加 GPG 密钥
    mkdir -p /etc/apt/keyrings
    curl -fsSL https://sing-box.app/gpg.key -o /etc/apt/keyrings/sagernet.asc
    chmod a+r /etc/apt/keyrings/sagernet.asc
    
    # 添加源
    cat > /etc/apt/sources.list.d/sagernet.sources << EOF
Types: deb
URIs: https://deb.sagernet.org/
Suites: *
Components: *
Enabled: yes
Signed-By: /etc/apt/keyrings/sagernet.asc
EOF
    
    # 安装
    apt update
    apt install -y systemd || true  # 确保 systemd 存在
    apt install -y sing-box
}

# 生成配置文件
generate_config() {
    local port=$1
    local email=$2
    local server_pass=$3
    local user_pass=$4
    
    log_info "生成配置文件..."
    
    mkdir -p /etc/sing-box
    
    cat > /etc/sing-box/config.json << EOF
{
  "log": {
    "level": "info"
  },
  "dns": {
    "servers": [
      {
        "address": "tls://8.8.8.8"
      }
    ]
  },
  "inbounds": [
    {
      "type": "shadowsocks",
      "listen": "::",
      "listen_port": ${port},
      "method": "2022-blake3-aes-256-gcm",
      "password": "${server_pass}",
      "users": [
        {
          "name": "${email}",
          "password": "${user_pass}"
        }
      ],
      "multiplex": {
        "enabled": true,
        "padding": true
      }
    }
  ],
  "outbounds": [
    {
      "type": "direct"
    }
  ],
  "route": {
    "rules": [
      {
        "protocol": "dns",
        "action": "hijack-dns"
      }
    ]
  }
}
EOF
    
    log_info "配置文件已保存到 /etc/sing-box/config.json"
}

# 生成 SS 链接
generate_ss_link() {
    local server_pass=$1
    local user_pass=$2
    local ip=$3
    local port=$4
    local email=$5
    
    # 格式: method:serverPassword:userPassword
    local credentials="2022-blake3-aes-256-gcm:${server_pass}:${user_pass}"
    local encoded=$(echo -n "$credentials" | base64 -w 0)
    
    echo "ss://${encoded}@${ip}:${port}?type=tcp#${email}"
}

# 主函数
main() {
    echo "=== sbctl - Sing-Box 一键安装脚本 ==="
    echo ""
    
    check_root
    
    # 获取用户输入
    read -p "请输入邮箱地址: " email
    if ! validate_email "$email"; then
        log_error "邮箱格式不正确"
        exit 1
    fi
    
    read -p "请输入监听端口 (1-65535): " port
    if ! validate_port "$port"; then
        log_error "端口号无效"
        exit 1
    fi
    
    echo ""
    log_info "开始安装..."
    echo ""
    
    # 执行安装步骤
    install_base_packages
    install_bbr
    install_singbox
    
    # 生成密码
    log_info "生成密码..."
    server_password=$(generate_password)
    user_password=$(generate_password)
    
    # 生成配置
    generate_config "$port" "$email" "$server_password" "$user_password"
    
    # 启动服务
    log_info "启动 sing-box 服务..."
    systemctl enable sing-box
    systemctl start sing-box || log_warn "服务启动失败（容器环境下正常）"
    
    # 获取公网 IP
    log_info "获取公网 IP..."
    public_ip=$(get_public_ip)
    
    # 生成链接
    log_info "生成链接..."
    ss_link=$(generate_ss_link "$server_password" "$user_password" "$public_ip" "$port" "$email")
    sub_content=$(echo -n "$ss_link" | base64 -w 0)
    
    # 输出结果
    echo ""
    echo "=== 安装完成 ==="
    echo ""
    echo "SS 链接:"
    echo "$ss_link"
    echo ""
    echo "Base64 订阅内容:"
    echo "$sub_content"
    echo ""
    echo "提示: 将 Base64 内容保存为文本文件，通过 HTTP 服务器提供订阅"
    echo ""
    
    # 询问是否重启
    echo "提示: BBR v3 需要重启服务器才能生效"
    read -p "是否现在重启服务器? (y/N): " answer
    
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        log_info "服务器将在 5 秒后重启..."
        sleep 5
        reboot
    else
        log_info "跳过重启。如需启用 BBR v3，请稍后手动重启: sudo reboot"
    fi
}

main "$@"
