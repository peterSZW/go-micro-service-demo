package main

import (
	//"github.com/xiaomi-tc/log15"
	"github.com/inconshreveable/log15"
	"net"
	"os"
)

func getHostIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops:" + err.Error())
		os.Exit(1)
	}
	for _, a := range addrs {
		//判断是否正确获取到IP
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalMulticast() && !ipnet.IP.IsLinkLocalUnicast() {
			//fmt.Println(ipnet.IP.String())
			log15.Info("getHostIP", "ip", ipnet)
			if ipnet.IP.To4() != nil {
				//os.Stdout.WriteString(ipnet.IP.String() + "\n")
				return ipnet.IP.String()
			}
		}
	}

	os.Stderr.WriteString("No Networking Interface Err!")
	log15.Error("getHostIP", "error", err)
	os.Exit(1)
	return ""
}
