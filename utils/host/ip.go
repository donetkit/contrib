package host

import (
	"net"
	"regexp"
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

func VerifyIpV4(ip string) bool {
	comp := regexp.MustCompile(`([1-9]?\d|1\d{2}|2[0-4]\d|25[0-5])(\.([1-9]?\d|1\d{2}|2[0-4]\d|25[0-5])){3}$`)
	subMatch := comp.FindAllStringSubmatch(ip, -1)
	if len(subMatch) <= 0 {
		return false
	}
	return true
}

func VerifyIpV6(ip string) bool {
	ip = strings.ReplaceAll(ip, "[", "")
	ip = strings.ReplaceAll(ip, "]", "")
	comp := regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:([0-9a-fA-F]{1,4}|:)|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]).){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]).){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)
	subMatch := comp.FindAllStringSubmatch(ip, -1)
	if len(subMatch) <= 0 {
		return false
	}
	return true
}
