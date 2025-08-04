// dot/client.go
package dot

import (
	"crypto/tls"
	customproxy "dns-forwarder/proxy"
	"fmt"
	"net"

	"github.com/miekg/dns"
	"golang.org/x/net/proxy"
)

// Client 定义了一个DoT客户端
type Client struct {
	Address string       // DoT服务器地址, 例如 "dot.pub:853"
	dialer  proxy.Dialer // 存储我们自定义的拨号器（可能带代理）
}

// NewClient 创建一个新的DoT客户端实例
// 如果提供了socks5Addr，它将通过SOCKS5代理路由所有DoT请求
func NewClient(serverAddress, socks5Addr string) (*Client, error) {
	// 使用我们的proxy模块来创建一个可能带有代理的拨号器
	dialer, err := customproxy.CreateDialer(socks5Addr)
	if err != nil {
		return nil, fmt.Errorf("无法创建DoT拨号器: %w", err)
	}

	return &Client{
		Address: serverAddress,
		dialer:  dialer,
	}, nil
}

// Resolve 发送DNS查询到DoT服务器并返回响应
// 这里我们手动处理连接过程，以支持自定义的代理拨号器
func (c *Client) Resolve(req *dns.Msg) (*dns.Msg, error) {
	// 1. 使用我们存储的拨号器建立TCP连接（此连接可能通过SOCKS5代理）
	conn, err := c.dialer.Dial("tcp", c.Address)
	if err != nil {
		return nil, fmt.Errorf("通过拨号器连接到 %s 失败: %w", c.Address, err)
	}
	defer conn.Close()

	// 2. 从地址中提取TLS服务器名 (Server Name Indication)
	host, _, err := net.SplitHostPort(c.Address)
	if err != nil {
		// 如果地址中没有端口，假定整个地址就是host
		host = c.Address
		// 在某些情况下，特别是当地址是IP时，可能会出现此错误。
		// 更好的做法是清理地址，但在这里我们假设它是一个有效的主机名或IP。
	}

	// 3. 在TCP连接之上进行TLS握手
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: host,
	})

	// 4. 将TLS连接包装在dns.Conn中，以便发送和接收DNS消息
	co := &dns.Conn{Conn: tlsConn}
	defer co.Close()

	// 5. 发送DNS请求
	if err := co.WriteMsg(req); err != nil {
		return nil, fmt.Errorf("向DoT服务器 %s 发送DNS请求失败: %w", c.Address, err)
	}

	// 6. 接收DNS响应
	resp, err := co.ReadMsg()
	if err != nil {
		return nil, fmt.Errorf("从DoT服务器 %s 读取DNS响应失败: %w", c.Address, err)
	}

	if resp == nil {
		return nil, fmt.Errorf("从DoT服务器 %s 收到空的响应", c.Address)
	}

	// 检查响应ID是否匹配请求ID
	if resp.Id != req.Id {
		return nil, fmt.Errorf("DoT响应ID不匹配: 收到 %d, 期望 %d", resp.Id, req.Id)
	}

	return resp, nil
}
