package host

import (
	"net"
	"strings"
)

// GetOutBoundIp 获取本地外网IP
func GetOutBoundIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err == nil {
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		ip := strings.Split(localAddr.String(), ":")[0]
		return ip
	}
	return GetLocalIp()
}

// GetLocalIp 获取本地IP
func GetLocalIp() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}
	return ""
}

// GetLocalIPs 获取本地IPs
func GetLocalIPs() []string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var ip []string
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			adds, _ := netInterfaces[i].Addrs()
			for _, address := range adds {
				if ips, ok := address.(*net.IPNet); ok && !ips.IP.IsLoopback() {
					if ips.IP.To4() != nil {
						ip = append(ip, ips.IP.String())
					}
				}
			}
		}
	}
	return ip
}
