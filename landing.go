package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// 从控制台读取本地端口号
	var localPort string
	fmt.Print("请输入落地端的本地端口号: ")
	fmt.Scanln(&localPort)

	// 启动落地端
	fmt.Printf("落地端启动，本地端口: %s\n", localPort)

	err := netx.ListenAndServe("tcp", ":"+localPort, func(local net.Conn) {
		fmt.Println("接收到连接:", local.RemoteAddr().String())

		// 从控制台读取目标IP和端口
		var targetIP, targetPort string
		fmt.Print("请输入目标IP地址: ")
		fmt.Scanln(&targetIP)
		fmt.Print("请输入目标端口号: ")
		fmt.Scanln(&targetPort)

		remote, err := net.Dial("tcp", targetIP+":"+targetPort)
		if err != nil {
			fmt.Println("连接目标服务器错误:", err)
			return
		}

		go io.Copy(remote, local)
		io.Copy(local, remote)

		_ = remote.Close()
		_ = local.Close()
	}, netx.WithReuseport())

	if err != nil {
		fmt.Println("启动落地端错误:", err)
		os.Exit(1)
	}
}
