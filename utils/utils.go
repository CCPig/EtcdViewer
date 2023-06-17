package utils

import (
	"fmt"
	"net"
	"time"
)

func CheckNet(host, port string) bool {
	timeout := time.Millisecond * 300
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		fmt.Printf("Failed to connect to %s:%s - %s\n", host, port, err.Error())
		return false
	}
	defer conn.Close()
	fmt.Printf("Connected to %s:%s\n", host, port)
	return true
}
