# sbctl - Sing-Box 一键安装脚本

一个简单的 Bash 脚本，用于自动化安装和配置 sing-box。

## 一键安装

**交互式安装（推荐）：**

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh)
```

或使用 wget：

```bash
bash <(wget -qO- https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh)
```

**非交互式安装：**

```bash
# 指定邮箱和端口
bash <(curl -fsSL https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh) -e user@example.com -p 8080
```

## 手动安装

如果你想先查看脚本内容再运行：

```bash
# 下载脚本
curl -O https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh

# 查看内容
cat install.sh

# 运行安装
chmod +x install.sh
sudo ./install.sh
```

## 使用步骤

1. 运行一键安装命令
2. 输入邮箱地址（用于标识）
3. 输入监听端口（1-65535）
4. 等待安装完成
5. 复制 SS 链接或订阅内容
6. 选择是否重启服务器（BBR v3 需要重启）

## 功能特性

- ✅ 自动安装 BBR v3（可选，失败不影响）
- ✅ 自动安装和配置 sing-box
- ✅ 自动生成安全的随机密码
- ✅ 生成 SS 链接和 Base64 订阅内容
- ✅ 自动获取服务器公网 IP
- ✅ 输入验证（邮箱、端口）
- ✅ 彩色日志输出
- ✅ 可选的服务器重启

## 安装步骤

脚本会自动执行以下步骤：

1. 检查 root 权限
2. 获取用户输入（邮箱、端口）
3. 安装基础工具（curl, vim）
4. 安装 BBR v3（可选）
5. 安装 sing-box
6. 生成随机密码
7. 创建配置文件
8. 启动服务
9. 生成 SS 链接和订阅内容
10. 询问是否重启服务器

## 配置文件

- 配置文件: `/etc/sing-box/config.json`
- 加密方法: `2022-blake3-aes-256-gcm`

## 服务管理

```bash
# 查看状态
sudo systemctl status sing-box

# 启动服务
sudo systemctl start sing-box

# 停止服务
sudo systemctl stop sing-box

# 重启服务
sudo systemctl restart sing-box

# 查看日志
sudo journalctl -u sing-box -f
```

## 系统要求

- Ubuntu 20.04+ 或 Debian 11+
- root 权限
- 网络连接

## 输出示例

```
=== 安装完成 ===

SS 链接:
ss://MjAyMi1ibGFrZTMtYWVzLTI1Ni1nY206...@1.2.3.4:8080?type=tcp#user@example.com

Base64 订阅内容:
c3M6Ly9NakF5TWkxaWJHRnJaVE10WVdWekxUSTFOaTFuWTIwNk...

提示: 将 Base64 内容保存为文本文件，通过 HTTP 服务器提供订阅
```

## 故障排查

### 端口被占用

```bash
# 检查端口占用
sudo netstat -tlnp | grep :8080

# 或使用 ss
sudo ss -tlnp | grep :8080
```

### 服务启动失败

```bash
# 查看详细日志
sudo journalctl -u sing-box -n 50

# 检查配置文件
sudo sing-box check -c /etc/sing-box/config.json
```

### BBR v3 安装失败

BBR v3 安装失败不会影响 sing-box 的正常使用，只是无法使用 BBR v3 拥塞控制算法。

## 卸载

```bash
# 停止并禁用服务
sudo systemctl stop sing-box
sudo systemctl disable sing-box

# 卸载 sing-box
sudo apt remove -y sing-box

# 删除配置文件
sudo rm -rf /etc/sing-box
sudo rm -f /etc/apt/sources.list.d/sagernet.sources
sudo rm -f /etc/apt/keyrings/sagernet.asc
```

## 许可证

MIT License
