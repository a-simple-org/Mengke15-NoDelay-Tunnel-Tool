package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"io"
	"net"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("1. 启动中转端")
	fmt.Println("2. 启动落地端")
	fmt.Println("3. 添加负载均衡IP地址")
	fmt.Println("4. 退出")

	for {
		fmt.Print("请选择功能: ")
		option, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入错误:", err)
			continue
		}

		option = strings.TrimSuffix(option, "\n")

		switch option {
		case "1":
			StartTransit()
		case "2":
			StartLanding()
		case "3":
			addIP(reader)
		case "4":
			fmt.Println("退出程序.")
			return
		default:
			fmt.Println("无效的选项，请重新选择.")
		}
	}
}

func addIP(reader *bufio.Reader) {
	fmt.Println("请输入要添加的IP地址和端口号 (格式: IP:端口)")
	ip, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取输入错误:", err)
		return
	}

	ip = strings.TrimSuffix(ip, "\n")

	// 将IP地址写入 ips.txt 文件中
	file, err := os.OpenFile("ips.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("打开文件错误:", err)
		return
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%s\n", ip)
	if err != nil {
		fmt.Println("写入文件错误:", err)
		return
	}

	fmt.Println("IP地址已添加.")
}

func StartLanding() {
	// 从控制台读取本地端口号
	var localPort string
	fmt.Print("请输入落地端的本地端口号: ")
	fmt.Scanln(&localPort)

	// 启动落地端
	fmt.Printf("落地端启动，本地端口: %s\n", localPort)

	l, err := net.Listen("tcp", ":"+localPort)
	if err != nil {
		fmt.Println("启动落地端错误:", err)
		os.Exit(1)
	}
	defer l.Close()

	for {
		local, err := l.Accept()
		if err != nil {
			fmt.Println("接受连接错误:", err)
			continue
		}
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
			continue
		}

		go io.Copy(remote, local)
		go io.Copy(local, remote)
	}
}

type LoadBalancer struct {
	ips []string
	idx int
}

func (lb *LoadBalancer) Next() string {
	if len(lb.ips) == 0 {
		return ""
	}
	ip := lb.ips[lb.idx]
	lb.idx = (lb.idx + 1) % len(lb.ips)
	return ip
}

func StartTransit() {
	// 从控制台读取本地端口号
	var localPort string
	fmt.Print("请输入中转端的本地端口号: ")
	fmt.Scanln(&localPort)

	// 加载负载均衡的IP列表
	ips, err := loadIPs()
	if err != nil {
		fmt.Println("加载IP列表错误:", err)
		return
	}

	if len(ips) == 0 {
		fmt.Println("IP列表为空，请先添加负载均衡IP地址.")
		return
	}

	// 创建负载均衡实例
	loadBalancer := &LoadBalancer{
		ips: ips,
		idx: 0,
	}

	// 使用负载均衡的 IP:端口 进行转发
	fmt.Printf("中转端启动，本地端口: %s\n", localPort)

	for {
		err := listenAndServe("tcp", ":"+localPort, func(local net.Conn) {
			server := loadBalancer.Next()
			remote, err := net.Dial("tcp", server)
			if err != nil {
				fmt.Println("连接远程服务器错误:", err)
				return
			}

			go io.Copy(remote, local)
			io.Copy(local, remote)

			_ = remote.Close()
			_ = local.Close()
		})

		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Println("启动中转端错误:", err)
			os.Exit(1)
		}

		time.Sleep(time.Second)
	}
}

func listenAndServe(network, address string, handler func(net.Conn)) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go handler(conn)
	}
}

func loadIPs() ([]string, error) {
	file, err := os.Open("ips.txt")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ips []string

	for scanner.Scan() {
		ip := scanner.Text()
		ips = append(ips, ip)
	}

	return ips, scanner.Err()
}
