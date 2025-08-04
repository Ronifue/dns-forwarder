// doh/client.go
package doh

import (
	"bytes"
	"dns-forwarder/proxy"
	"fmt"
	"io"
	"net/http"

	"github.com/miekg/dns"
)

// Client 定义了一个DoH客户端
type Client struct {
	URL        string
	httpClient *http.Client
}

// NewClient 创建一个新的DoH客户端实例
// 如果提供了socks5Addr，它将通过SOCKS5代理路由所有DoH请求
func NewClient(serverURL, socks5Addr string) (*Client, error) {
	// 使用我们的proxy模块来创建一个可能带有代理的http transport
	transport, err := proxy.NewHTTPTransport(socks5Addr)
	if err != nil {
		return nil, fmt.Errorf("无法创建DoH HTTP transport: %w", err)
	}

	return &Client{
		URL: serverURL,
		httpClient: &http.Client{
			Transport: transport,
		},
	}, nil
}

// Resolve 发送DNS查询到DoH服务器并返回响应
func (c *Client) Resolve(req *dns.Msg) (*dns.Msg, error) {
	// 将DNS消息打包成二进制格式
	packedReq, err := req.Pack()
	if err != nil {
		return nil, fmt.Errorf("打包DNS请求失败: %w", err)
	}

	// 创建HTTP POST请求
	httpReq, err := http.NewRequest("POST", c.URL, bytes.NewReader(packedReq))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/dns-message")
	httpReq.Header.Set("Accept", "application/dns-message")

	// 发送HTTP请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送DoH请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DoH服务器返回错误状态: %s", httpResp.Status)
	}

	// 读取响应体
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取DoH响应体失败: %w", err)
	}

	// 将二进制响应解包成DNS消息
	respMsg := new(dns.Msg)
	if err := respMsg.Unpack(body); err != nil {
		return nil, fmt.Errorf("解包DNS响应失败: %w", err)
	}

	return respMsg, nil
}
