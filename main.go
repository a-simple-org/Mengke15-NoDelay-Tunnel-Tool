package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
			startTransit()
		case "2":
			startLanding()
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

func startTransit() {
	// 在这里调用中转端的入口函数
}

func startLanding() {
	// 在这里调用落地端的入口函数
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
