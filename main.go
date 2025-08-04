// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"dns-forwarder/netutils" // <-- 修改导入路径
	"dns-forwarder/server"
)

func main() {
	// --- 参数定义 (增加-ecs) ---
	listenIP := flag.String("ip", "", "指定要监听的IP地址。留空以显示可选IP列表。")
	listenPort := flag.String("port", "53", "指定要监听的端口。")
	dohServer := flag.String("doh", "https://doh.pub/dns-query", "上游DoH服务器URL。")
	customECS := flag.String("ecs", "", "可选: 自定义ECS, 格式为IP (如8.8.8.8) 或CIDR (如8.8.8.0/24)。留空则自动检测公网IP。")
	listIPs := flag.Bool("list-ips", false, "仅列出本机所有可用的IPv4地址并退出。")

	flag.Parse()

	// --- 列出可选IP ---
	availableIPs, err := netutils.GetAvailableIPs()
	if err != nil {
		log.Fatalf("无法获取本地IP地址: %v", err)
	}

	if *listIPs {
		fmt.Println("本机可用的IPv4地址:")
		for _, ip := range availableIPs {
			fmt.Printf("- %s\n", ip.String())
		}
		os.Exit(0)
	}

	// --- 检查监听IP ---
	if *listenIP == "" {
		fmt.Println("错误: 未通过 -ip 参数指定监听IP。")
		fmt.Println("\n请从以下可用IP中选择一个:")
		for _, ip := range availableIPs {
			fmt.Printf("- %s\n", ip.String())
		}
		fmt.Printf("\n例如: go run . -ip %s\n", availableIPs[0].String())
		os.Exit(1)
	}

	// --- 启动服务器 ---
	listenAddr := fmt.Sprintf("%s:%s", *listenIP, *listenPort)
	if !strings.HasPrefix(*dohServer, "https://") {
		log.Fatalf("错误: DoH服务器地址必须以 https:// 开头。")
	}

	// 修改了Server的创建过程, 现在它可能返回error
	srv, err := server.NewServer(listenAddr, *dohServer, *customECS)
	if err != nil {
		log.Fatalf("初始化服务器失败: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务器时发生致命错误: %v", err)
	}
}
