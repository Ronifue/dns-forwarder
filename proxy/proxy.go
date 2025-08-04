// proxy/proxy.go
package proxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

// CreateDialer 创建一个网络拨号器。
// 如果提供了socks5Addr，返回的拨号器将通过SOCKS5代理建立连接。
// 否则，返回一个标准的直接连接拨号器。
// 返回的接口 proxy.Dialer 可以被DoT客户端用来建立底层TCP连接。
func CreateDialer(socks5Addr string) (proxy.Dialer, error) {
	// 基础拨号器，设置通用超时
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
	}

	// 如果没有提供代理地址，直接返回标准拨号器
	if socks5Addr == "" {
		return dialer, nil
	}

	log.Printf("网络: 检测到SOCKS5代理配置，将通过 %s 连接上游", socks5Addr)

	// proxy.SOCKS5函数需要一个实现了 proxy.Dialer 接口的 "forward" dialer。
	// &net.Dialer{} 实现了这个接口。
	// 该函数返回一个新的 proxy.Dialer 接口，它会先通过SOCKS5代理连接，然后再连接到目标地址。
	// 第三个参数 'auth' 在这里是nil，表示不需要认证。
	return proxy.SOCKS5("tcp", socks5Addr, nil, dialer)
}

// NewHTTPTransport 创建一个 *http.Transport。
// 如果提供了socks5Addr，该transport将被配置为使用SOCKS5代理。
// 这个transport专门给DoH客户端使用。
func NewHTTPTransport(socks5Addr string) (*http.Transport, error) {
	// 首先，利用CreateDialer获取一个可能带代理的拨号器
	dialer, err := CreateDialer(socks5Addr)
	if err != nil {
		return nil, err
	}

	// http.Transport 的 DialContext 字段需要一个特定签名的函数。
	// 我们需要将我们创建的 dialer 的 Dial 方法适配成这个签名。
	// 注意：我们的 dialer 没有 DialContext 方法，所以我们忽略传入的 context。
	// 在这个场景下通常是可接受的，因为超时已经在基础拨号器中设置了。
	dialContext := func(_ context.Context, network, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}

	// 创建并返回配置好的 transport
	return &http.Transport{
		// 使用我们自定义的拨号逻辑
		DialContext: dialContext,
		// 从环境中读取HTTP/HTTPS代理设置，这通常是良好的实践，
		// 但我们的DialContext会覆盖它，这里保留以备其他用途。
		Proxy: http.ProxyFromEnvironment,
		// 其他一些来自http.DefaultTransport的推荐设置
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}, nil
}
