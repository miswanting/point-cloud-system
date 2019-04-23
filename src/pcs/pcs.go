package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// 等待接入
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("123")
		// 在新的Go程里处理会话
		// 循环返回到等待新接入，就可以用协程处理接入
		go func(c net.Conn) {
			// 反弹
			c.Write([]byte("abc"))
			// 关闭链接
			c.Close()
		}(conn)
	}
}
