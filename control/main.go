package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("control start...")
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8964")
	if err != nil {
		fmt.Println("create udp addr err")
	}

	conn, _ := net.DialUDP("udp", nil, udpAddr)
	conn.Write([]byte("testing proxy control"))
	var task string
	for {
		fmt.Println("What task do you want to perform?(h for help)")
		fmt.Print("task:")
		fmt.Scan(&task)
		fmt.Printf("the task is %s. \n", task)
		switch task {
		case "h":
			help()
		case "c":
			fmt.Println("check all path status...")
			conn.Write([]byte("[task][check]"))
			status := make([]byte, 1024)
			n, _ := conn.Read(status)
			fmt.Printf("%s", status[:n])
		case "m":
			fmt.Println("migrate to specific path...")
			var pathStr string
			fmt.Print("Please enter the Path IP (local):")
			fmt.Scan(&pathStr)
			Command := ""
			Command += "[task][migra]"
			Command += pathStr
			conn.Write([]byte(Command))
			result := make([]byte, 1024)
			n, _ := conn.Read(result)
			fmt.Printf("result:%s\n", result[:n])
		default:
			fmt.Println("Unknown task!")
		}
		time.Sleep(time.Second)
	}
}

func help() {
	fmt.Println("c : check all path status")
	fmt.Println("m : migrate to specific path")
}
