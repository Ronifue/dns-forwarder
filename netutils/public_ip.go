// netutils/public_ip.go
package netutils

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetPublicIP 通过访问外部服务获取本机的公网IPv4地址。
func GetPublicIP(serviceURL string) (net.IP, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(serviceURL)
	if err != nil {
		return nil, fmt.Errorf("无法连接到IP查询服务 '%s': %w", serviceURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IP查询服务返回错误状态: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("无法读取IP查询服务响应: %w", err)
	}

	ipStr := strings.TrimSpace(string(body))
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("无法将 '%s' 解析为有效的IP地址", ipStr)
	}

	// 确保是IPv4
	if ip.To4() == nil {
		return nil, fmt.Errorf("获取到的公网IP不是IPv4地址: %s", ipStr)
	}

	return ip, nil
}
