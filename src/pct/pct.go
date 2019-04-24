package main

import (
	"log"
	"net"
	"bufio"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("[INFO]","正在拨号...")
	conn, err := net.Dial("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[FINE]","已通过",conn.LocalAddr(),"的地址连接到",conn.RemoteAddr(),"的主机！")
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readwriter := bufio.NewReadWriter(reader, writer)
	WriteString(readwriter,"123")
	fmt.Println("msg")
	msg := ReadString(readwriter)
	fmt.Println(msg)
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