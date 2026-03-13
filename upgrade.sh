#!/bin/bash

# sbctl 一键升级脚本
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/sail-tunnel/sbctl/main/upgrade.sh | bash

set -e

SBCTL_BIN="/usr/local/bin/sbctl"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

if [ "$EUID" -ne 0 ]; then
    log_error "请使用 root 权限运行: sudo bash upgrade.sh"
    exit 1
fi

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  SBCTL_ARCH="amd64" ;;
    aarch64) SBCTL_ARCH="arm64" ;;
    armv7l)  SBCTL_ARCH="arm"   ;;
    *)
        log_error "不支持的系统架构: $ARCH"
        exit 1
        ;;
esac

echo "=== sbctl 升级 ==="
echo ""

# 显示当前版本
if [ -f "$SBCTL_BIN" ]; then
    CURRENT=$("$SBCTL_BIN" --help 2>&1 | grep -oP 'sbctl \K[^\s]+' | head -1 || echo "未知")
    log_info "当前版本: $CURRENT"
else
    log_warn "未找到已安装的 sbctl，将执行全新安装"
fi

# 下载最新版本
DOWNLOAD_URL="https://github.com/sail-tunnel/sbctl/releases/latest/download/sbctl-linux-${SBCTL_ARCH}"
TMP_BIN=$(mktemp)

log_info "正在下载最新版本 (${SBCTL_ARCH})..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN"; then
    log_error "下载失败，请检查网络连接"
    rm -f "$TMP_BIN"
    exit 1
fi

chmod +x "$TMP_BIN"
mv "$TMP_BIN" "$SBCTL_BIN"

NEW=$("$SBCTL_BIN" --help 2>&1 | grep -oP 'sbctl \K[^\s]+' | head -1 || echo "未知")
log_info "升级完成，当前版本: $NEW"
echo ""
log_info "使用 'sbctl help' 查看所有命令"