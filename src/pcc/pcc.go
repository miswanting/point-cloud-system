package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"strconv"
)

const (
	PORT        = 80
	StopPattern = "\r\n\r\n"
)
func main() {
	currentPort:=PORT

	for {
		fmt.Println("[INFO]","正在尝试侦听",currentPort,"端口...")
		l, err := net.Listen("tcp", ":"+strconv.Itoa(currentPort))
		if err != nil {
			fmt.Println("[ERRO]",currentPort,"端口已被占用！")
			currentPort+=1
			continue
		}else {
			defer l.Close()
			for {
				// 等待接入
				fmt.Println("[INFO]","正在侦听",currentPort,"端口...")
				conn, err := l.Accept()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("[FINE]","已通过",conn.LocalAddr(),"的地址接收到来自",conn.RemoteAddr(),"的连接！")
				// 在新的Go程里处理会话
				// 循环返回到等待新接入，就可以用协程处理接入
				go func(c net.Conn) {
					reader := bufio.NewReader(conn)
					writer := bufio.NewWriter(conn)
					readwriter := bufio.NewReadWriter(reader, writer)
					msg := ReadString(readwriter)
					fmt.Println(msg)
					WriteString(readwriter,msg)
					c.Close()
					fmt.Println("[INFO]","连接已关闭！")
				}(conn)
			}
		}
	}
}

func Read(readwriter *bufio.ReadWriter)(b []byte)  {
	return
}
func Write(readwriter *bufio.ReadWriter,b []byte)  {
	
}
func ReadString(readwriter *bufio.ReadWriter)(str string)  {
	raw_msg, _ := readwriter.ReadString('\n')
	msg:=strings.Split(raw_msg, "\n")
	return msg[0]
}
func WriteString(readwriter *bufio.ReadWriter,str string)  {
	readwriter.WriteString(str+"\n")
	readwriter.Flush()
}