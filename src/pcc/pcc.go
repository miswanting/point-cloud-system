package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const (
	PORT        = 80
	StopPattern = "\r\n\r\n"
)

func main() {
	fmt.Println("正在建立主机...")
	l, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	fmt.Println("主机建立完成！")
	for {
		// 等待接入
		fmt.Println("正在等待连接...")
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("接收到连接！")
		// 在新的Go程里处理会话
		// 循环返回到等待新接入，就可以用协程处理接入
		go func(c net.Conn) {
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			readwriter := bufio.NewReadWriter(reader, writer)
			readwriter.Read()
			fmt.Println()
			c.Close()
			fmt.Println("连接已关闭！")
		}(conn)
	}
}
