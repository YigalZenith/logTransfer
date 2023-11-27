package utils

import (
	"fmt"
	"net"
)

// GetLocalIP 用来获取本地IP
func GetLocalIP() (addr string) {
	addrs, err := net.InterfaceAddrs()
	//fmt.Println(addrs)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				addr = ipnet.IP.String()
			}
		}
	}
	//fmt.Println(addr)
	return
}
