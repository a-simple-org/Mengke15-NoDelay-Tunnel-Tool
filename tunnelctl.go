package tunnelctl

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/abursavich/netx"
)

func main() {
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

	// 使用负载均衡的 IP:端口 进行转发
	fmt.Printf("中转端启动，本地端口: %s\n", localPort)

	for {
		err := netx.ListenAndServe("tcp", ":"+localPort, func(local net.Conn) {
			server := ips.Next().(string)
			remote, err := net.Dial("tcp", server)
			if err != nil {
				fmt.Println("连接远程服务器错误:", err)
				return
			}

			go io.Copy(remote, local)
			io.Copy(local, remote)

			_ = remote.Close()
			_ = local.Close()
		}, netx.WithReuseport())

		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Println("启动中转端错误:", err)
			os.Exit(1)
		}

		time.Sleep(time.Second)
	}
}

func loadIPs() (netx.WeightedList, error) {
	file, err := os.Open("ips.txt")
	if err != nil {
		if os.IsNotExist(err) {
			return netx.NewWeightedList(), nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ips := netx.NewWeightedList()

	for scanner.Scan() {
		ip := scanner.Text()
		ips.Add(ip, 1)
	}

	return ips, scanner.Err()
}
