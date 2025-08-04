// dot/client.go
package dot

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

// Client 定义了一个DoT客户端
type Client struct {
	Address    string       // DoT服务器地址, 例如 "dns.pub:853"
	dnsClient  *dns.Client  // 复用dns客户端以提高性能
}

// NewClient 创建一个新的DoT客户端实例
func NewClient(serverAddress string) *Client {
	return &Client{
		Address: serverAddress,
		dnsClient: &dns.Client{
			Net:          "tcp-tls", // 关键: 使用DNS-over-TLS
			Timeout:      5 * time.Second, // 设置一个合理的超时
		},
	}
}

// Resolve 发送DNS查询到DoT服务器并返回响应, 实现了server.Resolver接口
func (c *Client) Resolve(req *dns.Msg) (*dns.Msg, error) {
	// dns.Client的Exchange方法会处理连接、发送请求和接收响应的整个过程
	resp, _, err := c.dnsClient.Exchange(req, c.Address)
	if err != nil {
		return nil, fmt.Errorf("发送DoT请求到 %s 失败: %w", c.Address, err)
	}

	// 检查响应是否被截断。对于TCP/TLS来说，这通常不应该发生，但作为健壮性检查是好的。
	if resp.Truncated {
		// 在DoT中，截断的响应是一个异常情况，可能表示服务器端有问题。
		// 客户端可以选择重试，但在这里我们简单地返回错误。
		return nil, fmt.Errorf("从DoT服务器 %s 收到的响应被截断", c.Address)
	}

	if resp == nil {
		return nil, fmt.Errorf("从DoT服务器 %s 收到空的响应", c.Address)
	}

	return resp, nil
}
