# DNS Forwarder

一个轻量级的DNS转发服务器，旨在将本地的DNS请求安全地转发到上游的DNS-over-HTTPS (DoH)或DNS-over-TLS (DoT)服务器。

[English Version](README_en.md)

## 功能特性

- **多种上游协议**: 支持通过DoH和DoT两种安全协议将DNS请求加密发出。
- **SOCKS5代理支持**: 可以通过指定的SOCKS5代理连接上游服务器，保护您的网络足迹。
- **智能ECS支持**:
  - **自动检测**: 自动获取本机的公网IP地址，并将其用于EDNS客户端子网(ECS)选项，以获得更准确、更快速的CDN解析结果。
  - **自定义ECS**: 允许用户手动指定一个IP地址或CIDR作为ECS信息，适用于有特定网络需求的用户。
- **灵活的网络配置**:
  - 可以监听在任何指定的本地IP地址上。
  - 提供一个方便的工具，可以列出本机所有可用的IP地址，帮助用户选择。
- **跨平台与轻量级**: 使用Go语言编写，可以轻松交叉编译到各种操作系统（Windows, macOS, Linux等），并且运行时资源占用极低。

## 安装与构建

确保您的系统已经安装了Go语言环境（建议版本 1.18+）。

1.  **克隆或下载代码**:
    ```bash
    git clone <repository_url>
    cd dns-forwarder
    ```

2.  **构建可执行文件**:
    ```bash
    go build .
    ```
    执行成功后，当前目录下会生成一个名为 `dns-forwarder` (或 `dns-forwarder.exe`) 的可执行文件。

## 使用方法

该程序通过命令行参数进行配置。

### 参数说明

- `-ip <IP地址>`: **(必需)** 指定服务器监听的本地IP地址。
- `-port <端口>`: (可选) 指定监听的端口，默认为 `53`。
- `-doh <URL>`: (必需, 与-dot二选一) 指定上游DoH服务器的URL。例如: `https://doh.pub/dns-query`
- `-dot <地址>`: (必需, 与-doh二选一) 指定上游DoT服务器的地址和端口。例如: `dot.pub:853`
- `-socks5 <地址>`: (可选) 指定SOCKS5代理服务器的地址和端口。例如: `127.0.0.1:1080`
- `-ecs <IP/CIDR>`: (可选) 手动指定用于ECS的IP地址或CIDR。如果留空，程序将自动检测公网IP。
- `-ip-service <URL>`: (可选) 用于自动检测公网IP的查询服务地址，默认为 `https://4.ipw.cn`。
- `-list-ips`: (工具) 该参数会列出本机所有可用的IPv4地址，然后退出程序。这可以帮助您选择 `-ip` 参数的值。

### 使用示例

#### 示例 1: 列出可用的本地IP地址

在不确定应该监听哪个IP时，首先运行此命令：
```bash
./dns-forwarder -list-ips
```
输出可能如下:
```
本机可用的IPv4地址:
- 192.168.1.100
- 10.0.0.5
```

#### 示例 2: 使用DoH作为上游

选择一个IP (例如 `192.168.1.100`) 并使用公共DoH服务。
```bash
./dns-forwarder -ip 192.168.1.100 -doh https://dns.alidns.com/dns-query
```
服务器将在 `192.168.1.100:53` 上启动。将您的设备或路由器的DNS服务器设置为 `192.168.1.100` 即可开始使用。

#### 示例 3: 使用DoT并通过SOCKS5代理

假如您在本机 `1080` 端口有一个SOCKS5代理，并且想使用DoT服务。
```bash
./dns-forwarder -ip 192.168.1.100 -dot dns.google:853 -socks5 127.0.0.1:1080
```

#### 示例 4: 自定义ECS

如果您希望上游DNS服务器认为您的请求来自某个特定的IP地址（例如 `1.2.3.4`）。
```bash
./dns-forwarder -ip 192.168.1.100 -doh https://doh.pub/dns-query -ecs 1.2.3.4
```
或者使用一个CIDR范围：
```bash
./dns-forwarder -ip 192.168.1.100 -doh https://doh.pub/dns-query -ecs 1.2.3.0/24
```

## 注意事项

- 运行在Linux的53端口上可能需要 `root` 权限或 `CAP_NET_BIND_SERVICE` 权能。
- 必须在 `-doh` 和 `-dot` 中选择一个，且只能选择一个作为上游。
- DoH服务器的URL必须以 `https://` 开头。
