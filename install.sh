#!/bin/bash

# sbctl 安装引导脚本
# https://github.com/sail-tunnel/sbctl
#
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh | bash -s -- -r myserver -P shadowsocks

set -e

SBCTL_BIN="/usr/local/bin/sbctl"
SBCTL_RAW="https://raw.githubusercontent.com/sail-tunnel/sbctl/main/sbctl"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查 root
if [ "$EUID" -ne 0 ]; then
    log_error "请使用 root 权限运行: sudo bash install.sh"
    exit 1
fi

echo "=== sbctl 安装引导 ==="
echo ""

# ---------- 1. 基础工具 ----------
log_info "安装基础工具..."
apt update -qq
apt install -y curl vim openssl

# ---------- 2. BBR v3（可选） ----------
log_info "安装 BBR v3（可选）..."
if curl -L -s https://raw.githubusercontent.com/byJoey/Actions-bbr-v3/refs/heads/main/install.sh -o /tmp/bbr-install.sh 2>/dev/null; then
    echo "1" | bash /tmp/bbr-install.sh || log_warn "BBR v3 安装失败，跳过"
else
    log_warn "BBR v3 脚本下载失败，跳过"
fi

# ---------- 3. sing-box ----------
log_info "安装 sing-box..."
mkdir -p /etc/apt/keyrings
curl -fsSL https://sing-box.app/gpg.key -o /etc/apt/keyrings/sagernet.asc
chmod a+r /etc/apt/keyrings/sagernet.asc
cat > /etc/apt/sources.list.d/sagernet.sources <<EOF
Types: deb
URIs: https://deb.sagernet.org/
Suites: *
Components: *
Enabled: yes
Signed-By: /etc/apt/keyrings/sagernet.asc
EOF
apt update -qq
apt install -y systemd || true
apt install -y sing-box

# ---------- 4. 下载 sbctl ----------
log_info "下载 sbctl 管理工具..."
curl -fsSL "$SBCTL_RAW" -o "$SBCTL_BIN"
chmod +x "$SBCTL_BIN"
log_info "sbctl 已安装到 $SBCTL_BIN"

# ---------- 5. 初始化配置 ----------
log_info "开始初始化配置..."
echo ""
sbctl init "$@"

echo ""
log_info "全部完成！使用 'sbctl help' 查看所有命令。"
echo ""
echo "提示: BBR v3 需要重启服务器才能生效"
read -p "是否现在重启服务器? (y/N): " answer
if [[ "$answer" =~ ^[Yy]$ ]]; then
    log_info "服务器将在 5 秒后重启..."
    sleep 5
    reboot
else
    log_info "跳过重启。如需启用 BBR v3，请稍后手动重启: sudo reboot"
fi
