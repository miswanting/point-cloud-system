// 1.连接Point
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

func main() {
	id := getHash()
	log.Println(id)
	log.Println("[INFO]", "正在拨号...")
	conn, err := net.Dial("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[FINE]", conn.LocalAddr(), "<==>", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readwriter := bufio.NewReadWriter(reader, writer)
	WriteString(readwriter, "123")
	fmt.Println("msg")
	msg := ReadString(readwriter)
	fmt.Println(msg)
}

func Read(readwriter *bufio.ReadWriter) (b []byte) {
	return
}
func Write(readwriter *bufio.ReadWriter, b []byte) {

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
