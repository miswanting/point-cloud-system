// 1.建立Star
package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Version          = "0.1.0-190424"
	LogFile          = "pcs.log"
	ConfigFile       = "pcs.config.json"
	DefaultID        = "auto"
	DefaultProxyPort = 1994
)

var (
	logger    *log.Logger
	config    Config
	ID        string
	ProxyPort int
)

func main() {
	// 初始化
	os.Remove(LogFile) // 删除记录文件（如果有）
	// 指定记录文件
	logFile, err := os.OpenFile(LogFile, os.O_CREATE, 0777)
	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()
	// 记录文件和控制台双通
	w := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(w, "", log.LstdFlags)
	logger.Println("[HALO]", "Point Cloud System(PCS)", "[版本", Version+"]")
	logger.Println("[HALO]", "欢迎使用点云服务端！")
	// 处理配置文件
	config = Config{
		ID:        DefaultID,
		ProxyPort: DefaultProxyPort,
	}
	logger.Println("[INFO]", "正在查找配置文件...")
	if _, err := os.Stat(ConfigFile); err == nil {
		// 配置文件存在
		logger.Println("[INFO]", "正在加载配置文件...")
		f, err := os.Open(ConfigFile)
		if err != nil {
			logger.Fatal(err)
		}
		reader := bufio.NewReader(f)
		writer := bufio.NewWriter(f)
		readWriter := bufio.NewReadWriter(reader, writer)
		raw := ReadString(readWriter)
		json.Unmarshal([]byte(raw), &config)
		// logger.Println(string(config.ID))
	} else if os.IsNotExist(err) {
		// 配置文件不存在
		logger.Println("[INFO]", "正在创建配置文件...")
		j, _ := json.MarshalIndent(config, "", "    ")
		// logger.Println(string(j))
		f, err := os.Create(ConfigFile)
		if err != nil {
			logger.Fatal(err)
		}
		reader := bufio.NewReader(f)
		writer := bufio.NewWriter(f)
		readWriter := bufio.NewReadWriter(reader, writer)
		Write(readWriter, j)
	} else {
		logger.Fatal(err)
	}
	// 处理ID
	if config.ID == "auto" {
		ID = getHash()
	} else {
		ID = config.ID
	}
	logger.Println("[INFO]", "使用ID:", ID)
	startProxy()
}
func startProxy() {
	ProxyPort = config.ProxyPort
	for {
		logger.Println("[INFO]", "尝试在", ProxyPort, "端口搭建代理服务器...")
		l, err := net.Listen("tcp", ":"+strconv.Itoa(ProxyPort))
		if err != nil {
			logger.Println("[ERRO]", ProxyPort, "代理服务器端口被占用！")
			ProxyPort += 1
			continue
		} else {
			defer l.Close()
			for {
				// 等待接入
				logger.Println("[INFO]", "代理服务器侦听中...")
				conn, err := l.Accept()
				if err != nil {
					logger.Fatal(err)
				}
				logger.Println("[FINE]", "建立Star-Point连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
				// 在新的Go程里处理会话
				// 循环返回到等待新接入，就可以用协程处理接入
				go handleProxy(conn)
			}
		}
	}
}
func handleProxy(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readwriter := bufio.NewReadWriter(reader, writer)
	msg := ReadString(readwriter)
	for {
		if msg == "exit" {
			logger.Println("[INFO]", "Star-Point连接接收到关闭信号！")
			break
		} else {
			logger.Println("[TEST]", "接收到Point信息：", msg, "进行反弹测试！")
			WriteString(readwriter, msg)
		}
	}
	WriteString(readwriter, "exit")
	conn.Close()
	logger.Println("[INFO]", "Star-Point连接受控关闭！")
}
func Read(readWriter *bufio.ReadWriter) (p []byte) {
	// BUG
	_, err := readWriter.Read(p)
	if err != nil {
		logger.Fatal(err)
	}
	return p
}
func Write(readWriter *bufio.ReadWriter, p []byte) {
	readWriter.Write(p)
	readWriter.Flush()
}
func ReadString(readWriter *bufio.ReadWriter) (str string) {
	raw_msg, _ := readWriter.ReadString('\n')
	msg := strings.Split(raw_msg, "\n")
	return msg[0]
}
func WriteString(readWriter *bufio.ReadWriter, str string) {
	readWriter.WriteString(str + "\n")
	readWriter.Flush()
}
func getHash() (hash string) {
	salt := []byte(strconv.Itoa(rand.Int()) + strconv.FormatInt(time.Now().UnixNano(), 10))
	h := strings.ToUpper(fmt.Sprintf("%x", md5.Sum(salt)))
	return h
}

type Config struct {
	Mode      string // General Mode
	ID        string // Point ID ("auto"/)
	ProxyPort int    // Port for Data Transfer
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
