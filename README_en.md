# DNS Forwarder

A lightweight DNS forwarding server designed to securely forward local DNS requests to upstream DNS-over-HTTPS (DoH) or DNS-over-TLS (DoT) servers.

[中文文档](README.md)

## Features

- **Multiple Upstream Protocols**: Supports encrypting DNS requests using both DoH and DoT protocols.
- **SOCKS5 Proxy Support**: Can connect to upstream servers through a specified SOCKS5 proxy to protect your network footprint.
- **Intelligent ECS Support**:
  - **Auto-Detect**: Automatically fetches the host's public IP address and uses it for the EDNS Client Subnet (ECS) option to get more accurate and faster CDN query results.
  - **Custom ECS**: Allows users to manually specify an IP address or CIDR as ECS information, suitable for users with specific network requirements.
- **Flexible Network Configuration**:
  - Can listen on any specified local IP address.
  - Provides a convenient tool to list all available IP addresses on the host to help users make a selection.
- **Cross-Platform & Lightweight**: Written in Go, it can be easily cross-compiled for various operating systems (Windows, macOS, Linux, etc.) and has a very low runtime resource footprint.

## Installation and Build

Ensure you have a Go environment installed (version 1.18+ recommended).

1.  **Clone or download the code**:
    ```bash
    git clone <repository_url>
    cd dns-forwarder
    ```

2.  **Build the executable**:
    ```bash
    go build .
    ```
    Upon successful execution, an executable file named `dns-forwarder` (or `dns-forwarder.exe` on Windows) will be generated in the current directory.

## Usage

The program is configured via command-line arguments.

### Argument Details

- `-ip <IP Address>`: **(Required)** Specifies the local IP address for the server to listen on.
- `-port <Port>`: (Optional) Specifies the listening port. Defaults to `53`.
- `-doh <URL>`: (Required, choose one with -dot) Specifies the upstream DoH server URL. E.g., `https://doh.pub/dns-query`
- `-dot <Address>`: (Required, choose one with -doh) Specifies the upstream DoT server address and port. E.g., `dot.pub:853`
- `-socks5 <Address>`: (Optional) Specifies the SOCKS5 proxy server address and port. E.g., `127.0.0.1:1080`
- `-ecs <IP/CIDR>`: (Optional) Manually specifies the IP address or CIDR for ECS. If left empty, the program will auto-detect the public IP.
- `-ip-service <URL>`: (Optional) The query service URL for auto-detecting the public IP. Defaults to `https://4.ipw.cn`.
- `-list-ips`: (Utility) This flag lists all available IPv4 addresses on the host and then exits. It helps you choose a value for the `-ip` argument.

### Usage Examples

#### Example 1: List Available Local IP Addresses

When unsure which IP to listen on, run this command first:
```bash
./dns-forwarder -list-ips
```
The output might look like this:
```
Available IPv4 addresses on this machine:
- 192.168.1.100
- 10.0.0.5
```

#### Example 2: Use DoH as Upstream

Choose an IP (e.g., `192.168.1.100`) and use a public DoH service.
```bash
./dns-forwarder -ip 192.168.1.100 -doh https://dns.alidns.com/dns-query
```
The server will start on `192.168.1.100:53`. Set your device's or router's DNS server to `192.168.1.100` to start using it.

#### Example 3: Use DoT through a SOCKS5 Proxy

If you have a SOCKS5 proxy running on port `1080` of your local machine and want to use a DoT service.
```bash
./dns-forwarder -ip 192.168.1.100 -dot dns.google:853 -socks5 127.0.0.1:1080
```

#### Example 4: Custom ECS

If you want the upstream DNS server to think your request is coming from a specific IP address (e.g., `1.2.3.4`).
```bash
./dns-forwarder -ip 192.168.1.100 -doh https://doh.pub/dns-query -ecs 1.2.3.4
```
Or use a CIDR range:
```bash
./dns-forwarder -ip 192.168.1.100 -doh https://doh.pub/dns-query -ecs 1.2.3.0/24
```

## Notes

- Running on port 53 on Linux may require `root` privileges or the `CAP_NET_BIND_SERVICE` capability.
- You must choose one, and only one, from `-doh` and `-dot` as the upstream.
- The DoH server URL must start with `https://`.
