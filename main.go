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
	// --- 参数定义 ---
	listenIP := flag.String("ip", "", "指定要监听的IP地址。留空以显示可选IP列表。")
	listenPort := flag.String("port", "53", "指定要监听的端口。")
	dohServer := flag.String("doh", "", "上游DoH服务器URL (例如: https://doh.pub/dns-query)。")
	dotServer := flag.String("dot", "", "上游DoT服务器URL (例如: dot.pub:853)。")
	socks5Proxy := flag.String("socks5", "", "可选: SOCKS5代理服务器地址 (例如: 127.0.0.1:1080)。")
	customECS := flag.String("ecs", "", "可选: 自定义ECS, 格式为IP (如8.8.8.8) 或CIDR (如8.8.8.0/24)。留空则自动检测公网IP。")
	ipServiceURL := flag.String("ip-service", "https://4.ipw.cn", "可选: 用于自动检测公网IP的查询服务地址。")
	listIPs := flag.Bool("list-ips", false, "仅列出本机所有可用的IPv4地址并退出。")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: %s -ip <监听IP> [-port <端口>] [-doh <URL> | -dot <地址>] [-socks5 <地址>] [-ecs <IP/CIDR>] [-list-ips]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "参数:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\n注意: 必须在 -doh 和 -dot 中选择一个作为上游服务器。")
	}

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

	// --- 检查和验证参数 ---
	if *listenIP == "" {
		fmt.Println("错误: 未通过 -ip 参数指定监听IP。")
		fmt.Println("\n请从以下可用IP中选择一个:")
		for _, ip := range availableIPs {
			fmt.Printf("- %s\n", ip.String())
		}
		fmt.Printf("\n例如: go run . -ip %s -doh https://doh.pub/dns-query\n", availableIPs[0].String())
		os.Exit(1)
	}

	if (*dohServer == "" && *dotServer == "") || (*dohServer != "" && *dotServer != "") {
		log.Println("错误: 您必须在 -doh 和 -dot 中选择一个，且只能选择一个。")
		flag.Usage()
		os.Exit(1)
	}

	if *dohServer != "" && !strings.HasPrefix(*dohServer, "https://") {
		log.Fatalf("错误: DoH服务器地址 '%s' 必须以 https:// 开头。", *dohServer)
	}

	// --- 启动服务器 ---
	listenAddr := fmt.Sprintf("%s:%s", *listenIP, *listenPort)

	// 根据新的NewServer函数签名来创建实例
	srv, err := server.NewServer(listenAddr, *dohServer, *dotServer, *socks5Proxy, *customECS, *ipServiceURL)
	if err != nil {
		log.Fatalf("初始化服务器失败: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务器时发生致命错误: %v", err)
	}
}
