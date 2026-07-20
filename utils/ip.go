package utils

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/ipipdotnet/ipdb-go"
)

// GetIPAddr 获取客户端 IP 地址
// 它会检查 X-Forwarded-For 和 X-Real-IP HTTP 头部，
// 如果这些头部不存在，则回退到请求的 RemoteAddr。
func GetIPAddr(r *http.Request) string {
	// 尝试从 X-Forwarded-For 获取 IP
	// X-Forwarded-For 可以是一个逗号分隔的 IP 列表 (client, proxy1, proxy2)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// 取列表中的第一个 IP，它通常是客户端的真实 IP
		parts := strings.Split(ip, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// 确保获取的是有效的 IP 地址
		for _, part := range parts {
			if net.ParseIP(part) != nil {
				return part
			}
		}
	}

	// 尝试从 X-Real-IP 获取 IP
	ip = r.Header.Get("X-Real-IP")
	if ip != "" && net.ParseIP(ip) != nil {
		return ip
	}

	// 回退到 RemoteAddr
	// RemoteAddr 的格式可能是 "IP:port"
	ipHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(ipHost) != nil {
		return ipHost
	}

	// 如果 SplitHostPort 失败（例如 RemoteAddr 只有 IP），直接返回 RemoteAddr
	// 但也需要验证它是否是有效的 IP
	if net.ParseIP(r.RemoteAddr) != nil {
		return r.RemoteAddr
	}

	return ""
}

var (
	cityDB   *ipdb.City
	ipdbOnce sync.Once
	ipdbErr  error
)

// initIPDB 初始化 IPDB 数据库（单例模式）
func initIPDB() {
	ipdbOnce.Do(func() {
		dbPath := "assets/qqwry.ipdb"
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			if _, err2 := os.Stat("asset/qqwry.ipdb"); err2 == nil {
				dbPath = "asset/qqwry.ipdb"
			}
		}

		cityDB, ipdbErr = ipdb.NewCity(dbPath)
		if ipdbErr != nil {
			Error("无法加载 IPDB 数据库: " + ipdbErr.Error())
		} else {
			Info("IPDB 数据库已加载 (" + dbPath + ")")
		}
	})
}

// isPrivateIP 判断是否为局域网 IP 地址（包括 IPv4 和 IPv6）
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 私有地址范围
	// 10.0.0.0/8
	// 172.16.0.0/12
	// 192.168.0.0/16
	// 127.0.0.0/8 (loopback)
	// 169.254.0.0/16 (link-local)
	privateIPv4Blocks := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(169, 254, 0, 0), Mask: net.CIDRMask(16, 32)},
	}

	// 检查 IPv4 私有地址
	if parsedIP.To4() != nil {
		for _, block := range privateIPv4Blocks {
			if block.Contains(parsedIP) {
				return true
			}
		}
		return false
	}

	// IPv6 私有地址范围
	// ::1/128 (loopback)
	// fc00::/7 (unique local address)
	// fe80::/10 (link-local)
	// ff00::/8 (multicast)
	if parsedIP.IsLoopback() {
		return true
	}
	if parsedIP.IsLinkLocalUnicast() {
		return true
	}
	if parsedIP.IsLinkLocalMulticast() {
		return true
	}
	// 检查 IPv6 unique local address (fc00::/7)
	if len(parsedIP) == net.IPv6len && parsedIP[0] >= 0xfc && parsedIP[0] <= 0xfd {
		return true
	}

	return false
}

// GetIPLocation 获取 IP 地址的位置信息
// 返回格式：国家-省-市-区，局域网地址返回"局域网"
func GetIPLocation(ip string) string {
	if ip == "" {
		return ""
	}

	// 判断是否为局域网地址
	if isPrivateIP(ip) {
		return "局域网"
	}

	// 确保 IPDB 已初始化
	initIPDB()

	// 如果初始化失败，返回空字符串
	if ipdbErr != nil || cityDB == nil {
		return ""
	}

	// 查询 IP 地址信息
	info, err := cityDB.FindInfo(ip, "CN")
	if err != nil {
		return ""
	}

	// 构建返回字符串：国家-省-市-区
	var parts []string
	if info.CountryName != "" {
		parts = append(parts, info.CountryName)
	}
	if info.RegionName != "" {
		parts = append(parts, info.RegionName)
	}
	if info.CityName != "" {
		parts = append(parts, info.CityName)
	}
	if info.DistrictName != "" {
		parts = append(parts, info.DistrictName)
	}

	return strings.Join(parts, "-")
}
