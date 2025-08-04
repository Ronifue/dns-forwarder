// server/server.go
package server

import (
	"dns-forwarder/doh"
	"dns-forwarder/dot"
	"dns-forwarder/netutils" // <-- 修改导入路径
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"
)

// Resolver 定义了所有上游DNS解析器（DoH, DoT等）都必须实现的通用接口。
// 这使得服务器的核心逻辑可以与具体的解析器实现解耦。
type Resolver interface {
	// Resolve 接收一个DNS查询消息，并返回一个响应消息或一个错误。
	Resolve(req *dns.Msg) (*dns.Msg, error)
}

// Server 结构体持有服务器的配置和预计算的ECS信息
type Server struct {
	listenAddr string
	resolver   Resolver // 使用接口代替具体实现
	ecsIP      net.IP   // 预先计算好的ECS IP
	ecsNetmask uint8    // 预先计算好的ECS子网掩码
}

// NewServer 创建一个新的DNS服务器实例
// 它现在可以根据参数选择DoH或DoT作为上游解析器
func NewServer(listenAddr, dohServerURL, dotServerURL, customECS, ipServiceURL string) (*Server, error) {
	// --- Part 1: 选择和配置上游解析器 ---
	var resolver Resolver
	if dohServerURL != "" {
		log.Printf("使用上游 DoH 服务器: %s", dohServerURL)
		resolver = doh.NewClient(dohServerURL)
	} else if dotServerURL != "" {
		log.Printf("使用上游 DoT 服务器: %s", dotServerURL)
		resolver = dot.NewClient(dotServerURL)
	} else {
		// 这本应在main.go中被阻止，但作为安全检查
		return nil, fmt.Errorf("必须提供一个DoH或DoT服务器地址")
	}

	// --- Part 2: 处理ECS逻辑 (这部分保持不变) ---
	var ecsIP net.IP
	var ecsNetmask uint8 = 24 // 默认IPv4子网掩码

	if customECS != "" {
		// 情况1: 用户自定义了ECS
		log.Printf("使用自定义ECS: %s", customECS)
		var err error
		_, ipNet, err := net.ParseCIDR(customECS)
		if err != nil {
			ecsIP = net.ParseIP(customECS)
			if ecsIP == nil {
				return nil, fmt.Errorf("无效的自定义ECS格式: '%s', 请使用IP或CIDR格式", customECS)
			}
		} else {
			ecsIP = ipNet.IP
			maskSize, _ := ipNet.Mask.Size()
			ecsNetmask = uint8(maskSize)
		}
	} else {
		// 情况2: 自动检测公网IP
		log.Println("正在自动检测公网IP用于ECS...")
		var err error
		ecsIP, err = netutils.GetPublicIP(ipServiceURL)
		if err != nil {
			return nil, fmt.Errorf("自动获取公网IP失败: %w", err)
		}
		log.Printf("成功检测到公网IP: %s, 将使用 %s/%d 作为ECS", ecsIP, ecsIP, ecsNetmask)
	}

	// --- Part 3: 创建服务器实例 ---
	return &Server{
		listenAddr: listenAddr,
		resolver:   resolver, // 存储接口
		ecsIP:      ecsIP,
		ecsNetmask: ecsNetmask,
	}, nil
}

// Start 方法保持不变...
func (s *Server) Start() error {
	handler := dns.NewServeMux()
	handler.HandleFunc(".", s.handleRequest)
	go s.startListener("udp", handler)
	go s.startListener("tcp", handler)
	log.Printf("DNS转发服务器已在 %s 上启动...", s.listenAddr)
	select {}
}

func (s *Server) startListener(netType string, handler *dns.ServeMux) {
	srv := &dns.Server{Addr: s.listenAddr, Net: netType, Handler: handler}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("无法在 %s 上启动 %s 监听器: %s\n", s.listenAddr, netType, err)
	}
}

// handleRequest 现在通过通用的resolver接口转发请求
func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	forwardReq := r.Copy()

	// 不论客户端是否支持EDNS, 我们都为转发的请求添加EDNS和ECS
	opt := forwardReq.IsEdns0()
	if opt == nil {
		opt = new(dns.OPT)
		opt.Hdr.Name = "."
		opt.Hdr.Rrtype = dns.TypeOPT
		forwardReq.Extra = append(forwardReq.Extra, opt)
	}

	var ecs *dns.EDNS0_SUBNET
	// 查找现有的ECS选项
	for _, option := range opt.Option {
		if e, ok := option.(*dns.EDNS0_SUBNET); ok {
			ecs = e
			break
		}
	}

	if ecs == nil {
		// 如果不存在, 创建并附加一个新的ECS选项
		ecs = new(dns.EDNS0_SUBNET)
		ecs.Code = dns.EDNS0SUBNET
		opt.Option = append(opt.Option, ecs)
	}

	// 使用服务器启动时确定的ECS信息填充或覆盖
	ecs.Family = 1 // 1 for IPv4
	ecs.SourceNetmask = s.ecsNetmask
	ecs.SourceScope = 0
	ecs.Address = s.ecsIP.To4() // 确保是4字节的IPv4地址

	clientIP := w.RemoteAddr().(*net.UDPAddr).IP
	log.Printf("来自 %s 的请求: %s, 附加ECS: %s/%d", clientIP.String(), r.Question[0].String(), s.ecsIP.String(), s.ecsNetmask)

	// 使用s.resolver接口进行解析，无需关心具体是DoH还是DoT
	resp, err := s.resolver.Resolve(forwardReq)
	if err != nil {
		log.Printf("错误: 转发DNS请求失败: %v", err)
		dns.HandleFailed(w, r)
		return
	}

	resp.Id = r.Id
	if err := w.WriteMsg(resp); err != nil {
		log.Printf("错误: 向客户端 %s 写回响应失败: %v", clientIP.String(), err)
	}
}
