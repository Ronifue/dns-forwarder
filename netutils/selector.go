// netutilss/selector.go
package netutils

import (
	"fmt"
	"net"
)

// GetAvailableIPs 扫描本机所有网络接口，返回一个可用的非环回IPv4地址列表。
func GetAvailableIPs() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("无法获取网络接口: %w", err)
	}

	var availableIPs []net.IP
	for _, i := range ifaces {
		// 忽略无效、环回和点对点接口
		if (i.Flags&net.FlagUp == 0) || (i.Flags&net.FlagLoopback != 0) || (i.Flags&net.FlagPointToPoint != 0) {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			return nil, fmt.Errorf("无法获取接口 %s 的地址: %w", i.Name, err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 我们只关心IPv4地址，因为这是局域网设备最常用的
			if ip != nil && ip.To4() != nil {
				availableIPs = append(availableIPs, ip)
			}
		}
	}

	if len(availableIPs) == 0 {
		return nil, fmt.Errorf("未找到可用的IPv4地址")
	}

	return availableIPs, nil
}
