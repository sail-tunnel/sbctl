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

# 检测系统架构
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  
        SBCTL_ARCH="amd64"
        ;;
    aarch64) 
        SBCTL_ARCH="arm64"
        ;;
    armv7l)  
        SBCTL_ARCH="arm"
        ;;
    *)       
        log_error "不支持的系统架构: $ARCH"
        exit 1
        ;;
esac

echo "=== sbctl 安装引导 ==="
echo "系统架构: $ARCH (sbctl: $SBCTL_ARCH)"
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
log_info "安装 sing-box v1.13.0..."

# 使用官方安装脚本安装指定版本
SING_VERSION="1.13.0"
if curl -fsSL https://sing-box.app/deb-install.sh | bash -s -- v${SING_VERSION}; then
    log_info "sing-box v${SING_VERSION} 安装完成"
else
    log_error "sing-box 安装失败"
    exit 1
fi

# 启用服务
systemctl enable sing-box
log_info "sing-box 服务已启用"

# ---------- 4. 安装 sbctl ----------
log_info "获取 sbctl 管理工具 (Go Edition)..."

# 假设 Release 最终发布的资产名为: sbctl-linux-amd64
SBCTL_BIN_URL="https://github.com/sail-tunnel/sbctl/releases/latest/download/sbctl-linux-${SBCTL_ARCH}"

if ! curl -fsSL "$SBCTL_BIN_URL" -o "$SBCTL_BIN"; then
    log_warn "从 Release 下载二进制失败，尝试回退使用 Bash 版本或本地构建。"
    # 此处作为兜底，如果用户克隆了仓库并运行本地 install.sh
    if [ -f "./sbctl-linux-${SBCTL_ARCH}" ]; then
        cp "./sbctl-linux-${SBCTL_ARCH}" "$SBCTL_BIN"
    else
        # 兜底层级2：这里应该处理成使用旧版的 bash 或者提示用户编译 (暂时省略展开)
        log_error "无法获取 sbctl 二进制文件，请检查网络或构建环境。"
        exit 1
    fi
fi

chmod +x "$SBCTL_BIN"
log_info "sbctl (${SBCTL_ARCH}) 已安装到 $SBCTL_BIN"

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
