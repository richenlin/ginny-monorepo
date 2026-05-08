package ip

import (
	"net"
	"net/http"
	"strings"
)

// GetLocalIP4 gets local ip address.
func GetLocalIP4() (ip string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func isIntranetIpv4(ip string) bool {
	if strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "169.254.") ||
		strings.HasPrefix(ip, "172.") ||
		strings.HasPrefix(ip, "10.") {
		return true
	}
	return false
}

// GetAvailablePort returns a port at random
func GetAvailablePort() int {
	l, _ := net.Listen("tcp", ":0") // listen on localhost
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	return port
}

// GetIPFromMeta returns IP address from request.
// Only when it used use proxy
func GetIPFromMeta(mds map[string]string) string {
	if ip := mds["x-forwarded-for"]; ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

// GetIPFromHTTPRequest get ip from http request
func GetIPFromHTTPRequest(r *http.Request) string {
	ip := r.Header.Get("x-forwarded-for")
	if ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	r.Header.Set("x-forwarded-for", host)
	return host
}

// GetRemoteIP get ip from http request
func GetRemoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
