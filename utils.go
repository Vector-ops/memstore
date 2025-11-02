package main

import (
	"fmt"
	"net"
)

func formatHostPort(host, port string) string {
	ip := net.ParseIP(host)
	if ip != nil && ip.To4() == nil {
		return fmt.Sprintf("[%s]:%s", host, port)
	}
	return fmt.Sprintf("%s:%s", host, port)
}
