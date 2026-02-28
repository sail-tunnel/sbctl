# sbctl - Sing-Box 管理工具 (Go 版本)

`sbctl` 是一个功能强大、易于使用的 Sing-Box 服务管理工具。它支持多协议共存、批量用户管理，并且可以自动生成所有客户端连接配置和订阅。

## 功能特性

- ✅ **多协议支持**：支持 Shadowsocks 2022、VMess + WebSocket、Hysteria2、WireGuard。
- ✅ **多端口共存**：可以在不同的端口上同时运行多种协议，互不干扰。
- ✅ **多用户管理**：支持向已有的协议端口动态添加新用户/客户端。
- ✅ **自动配置生成**：自动生成带备注的分享链接、Base64 订阅以及 WireGuard 客户端配置文件。
- ✅ **无依赖运行**：Go 语言编写，编译为单一二进制文件，部署简单，无需系统安装 Python 或其它依赖。
- ✅ **自动优化**：集成了 BBR v3 加速安装。

## 快速开始

### 1. 一键安装

使用 **curl**：

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh)
```

或使用 **wget**：

```bash
bash <(wget -qO- https://raw.githubusercontent.com/sail-tunnel/sbctl/main/install.sh)
```

### 2. 初始化第一个节点

```bash
# 启动交互式安装
sbctl init

# 或非交互式一键设置 (Shadowsocks 示例)
sbctl init -r MyServer1 -p 8080 -P shadowsocks
```

### 3. 添加更多节点/协议

`sbctl` 的 `init` 命令非常智能，如果你运行它时指定的是一个新端口，它会自动将新协议追加到现有配置中，而不是覆盖。

```bash
# 在 8081 端口追加一个 Hysteria2 节点
sbctl init -r Hy2Server -p 8081 -P hysteria2
```

### 4. 添加多设备/多用户

你可以向任何已经存在的端口追加新用户：

```bash
# 给 8080 端口添加一个名为 "Phone" 的新客户端
sbctl add-user -p 8080 -r Phone
```

## 命令说明

| 命令 | 说明 |
| :--- | :--- |
| `sbctl init` | 初始化或追加新的协议监听节点 |
| `sbctl add-user` | 向已有端口追加新客户端(用户) |
| `sbctl show` | 显示当前所有节点的连接信息和订阅 |
| `sbctl show -p <port>` | 仅展示特定端口下的节点/用户详情 |
| `sbctl status` | 查看 sing-box 运行状态 |
| `sbctl restart` | 重启服务使新配置生效 (通常 add/init 会自动执行) |
| `sbctl log` | 查看服务运行日志 |

## 本地构建

如果你想自己从源码编译：

```bash
# 需开启 Go 1.21+
go mod download
go build -o sbctl ./cmd/sbctl
```

## 许可证

MIT License
