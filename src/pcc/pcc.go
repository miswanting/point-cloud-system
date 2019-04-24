// 1.建立Point
// 2.连接Star
// 3.连接Point
// 4.连接Cloud
package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	PORT        = 80
	StopPattern = "\r\n\r\n"
)

type Config struct {
	ID   string
	Port int
}
type PointInfo struct {
	ID         string
	LocalAddr  string
	GlobalAddr string
	Neighbors  []string
}
type CloudInfo struct {
	ID     string
	Points []PointInfo
}
type StarInfo struct {
	ID         string
	LocalAddr  string
	GlobalAddr string
	Clouds     []CloudInfo
}

func main() {
	// 初始化
	// id:=getHash()
	currentPort := PORT
	for {
		log.Println("[INFO]", "正在尝试侦听", currentPort, "端口...")
		l, err := net.Listen("tcp", ":"+strconv.Itoa(currentPort))
		if err != nil {
			log.Println("[ERRO]", currentPort, "端口已被占用！")
			currentPort += 1
			continue
		} else {
			defer l.Close()
			for {
				// 等待接入
				log.Println("[INFO]", "正在侦听", currentPort, "端口...")
				conn, err := l.Accept()
				if err != nil {
					log.Fatal(err)
				}
				log.Println("[FINE]", conn.LocalAddr(), "<==>", conn.RemoteAddr())
				// 在新的Go程里处理会话
				// 循环返回到等待新接入，就可以用协程处理接入
				go func(c net.Conn) {
					reader := bufio.NewReader(conn)
					writer := bufio.NewWriter(conn)
					readwriter := bufio.NewReadWriter(reader, writer)
					msg := ReadString(readwriter)
					fmt.Println(msg)
					WriteString(readwriter, msg)
					c.Close()
					log.Println("[INFO]", "连接已关闭！")
				}(conn)
			}
		}
	}
}

func Read(readwriter *bufio.ReadWriter) (p []byte) {
	return
}
func Write(readwriter *bufio.ReadWriter, p []byte) {

}
func ReadString(readwriter *bufio.ReadWriter) (str string) {
	raw_msg, _ := readwriter.ReadString('\n')
	msg := strings.Split(raw_msg, "\n")
	return msg[0]
}
func WriteString(readwriter *bufio.ReadWriter, str string) {
	readwriter.WriteString(str + "\n")
	readwriter.Flush()
}
func getHash() (hash string) {
	salt := []byte(strconv.Itoa(rand.Int()) + strconv.FormatInt(time.Now().UnixNano(), 10))
	h := strings.ToUpper(fmt.Sprintf("%x", md5.Sum(salt)))
	return h
}
