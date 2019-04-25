// 1.建立Point
// 2.连接Star
// 3.连接Point
// 4.连接Cloud
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
	Version            = "0.1.0-190424"
	LogFile            = "pcc.log"
	ConfigFile         = "pcc.config.json"
	DefaultID          = "auto"
	DefaultMonitorPort = 80
	DefaultProxyPort   = 1994
	DefaultStarAddress = "localhost"
	DefaultStarPort    = 1995
)

var (
	logger      *log.Logger
	config      Config
	ID          string
	MonitorPort int
	ProxyPort   int
	StarRW      *bufio.ReadWriter
	AppRW       *bufio.ReadWriter
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
	logger.Println("[HALO]", "欢迎使用点云客户端！")
	// 处理配置文件
	config = Config{
		ID:          DefaultID,
		MonitorPort: DefaultMonitorPort,
		ProxyPort:   DefaultProxyPort,
		StarAddress: DefaultStarAddress,
		StarPort:    DefaultStarPort,
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
	// 尝试连接Star
	go connectToStar()
	// 启动代理服务器
	go startProxy()
	startMonitor()
}
func startMonitor() {
	MonitorPort := config.MonitorPort
	for {
		logger.Println("[INFO]", "尝试在", MonitorPort, "端口搭建控制台...")
		l, err := net.Listen("tcp", ":"+strconv.Itoa(MonitorPort))
		if err != nil {
			logger.Println("[ERRO]", MonitorPort, "控制台被占用！")
			MonitorPort += 1
			continue
		} else {
			defer l.Close()
			for {
				// 等待接入
				logger.Println("[INFO]", "控制台侦听中...")
				conn, err := l.Accept()
				if err != nil {
					logger.Fatal(err)
				}
				logger.Println("[FINE]", "建立Point-Monitor连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
				// 在新的Go程里处理会话
				// 循环返回到等待新接入，就可以用协程处理接入
				go handleMonitor(conn)
			}
		}
	}
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
				logger.Println("[FINE]", "建立Point-App连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
				// 在新的Go程里处理会话
				// 循环返回到等待新接入，就可以用协程处理接入
				go handleProxy(conn)
			}
		}
	}
}
func connectToStar() {
	logger.Println("[INFO]", "正在尝试连接点云服务器", config.StarAddress+":"+strconv.Itoa(config.StarPort), "...")
	conn, err := net.Dial("tcp", config.StarAddress+":"+strconv.Itoa(config.StarPort))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[FINE]", "建立Point-Star连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	StarRW = bufio.NewReadWriter(reader, writer)
	m := map[string]interface{}{
		"CMD": "signup",
		"ID":  ID,
	}
	WriteMap(StarRW, m)
	for {
		m = ReadMap(StarRW)
		if m["CMD"] == "close" {
			logger.Println("[INFO]", "Point-Star连接接收到关闭信号！")
			break
		} else {
			logger.Println("[INFO]", "接收到Star请求：", m["CMD"])
			if m["CMD"] == "login" {
				WriteMap(AppRW, m)
			}
		}
	}
	m = map[string]interface{}{
		"CMD": "close",
	}
	WriteMap(StarRW, m)
	conn.Close()
	logger.Println("[INFO]", "Point-Star连接受控关闭！")
}
func handleMonitor(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readwriter := bufio.NewReadWriter(reader, writer)
	msg := ReadString(readwriter)
	fmt.Println(msg)
	WriteString(readwriter, msg)
	conn.Close()
	logger.Println("[INFO]", "连接已关闭！")
}
func handleProxy(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	AppRW = bufio.NewReadWriter(reader, writer)
	for {
		m := ReadMap(AppRW)
		if m["CMD"] == "close" {
			logger.Println("[INFO]", "Point-App连接接收到关闭信号！")
			break
		} else {
			logger.Println("[INFO]", "接收到App请求：", m["CMD"])
			if m["CMD"] == "login" {
				WriteMap(StarRW, m)
			}
		}
	}
	m := map[string]interface{}{
		"CMD": "close",
	}
	WriteMap(AppRW, m)
	conn.Close()
	logger.Println("[INFO]", "Point-App连接受控关闭！")
}
func Str2Map(s string) (m map[string]interface{}) {
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		logger.Fatal(err)
	}
	return m
}
func Map2Str(m map[string]interface{}) (s string) {
	b, err := json.Marshal(m)
	if err != nil {
		logger.Fatal(err)
	}
	return string(b)
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
	raw_msg, err := readWriter.ReadString('\n')
	if err != nil {
		logger.Fatal("[WARN]", err)
		return "exit"
	}
	msg := strings.Split(raw_msg, "\n")
	return msg[0]
}
func WriteString(readWriter *bufio.ReadWriter, str string) {
	_, err := readWriter.WriteString(str + "\n")
	if err != nil {
		logger.Fatal("[WARN]", err)
	}
	readWriter.Flush()
}
func ReadMap(readWriter *bufio.ReadWriter) (m map[string]interface{}) {
	msg := ReadString(readWriter)
	return Str2Map(msg)
}
func WriteMap(readWriter *bufio.ReadWriter, m map[string]interface{}) {
	WriteString(readWriter, Map2Str(m))
}
func getHash() (hash string) {
	salt := []byte(strconv.Itoa(rand.Int()) + strconv.FormatInt(time.Now().UnixNano(), 10))
	h := strings.ToUpper(fmt.Sprintf("%x", md5.Sum(salt)))
	return h
}

type Config struct {
	Mode        string // General Mode
	ID          string // Point ID ("auto"/)
	Username    string // Username For Star/Cloud Authentication
	Password    string // Password For Star/Cloud Authentication
	MonitorPort int    // Port for Monitoring PCS Runtime Data
	ProxyPort   int    // Port for Data Transfer
	StarAddress string // IP address for Point Cloud Server you wish to connect to
	StarPort    int    // Port for Point Cloud Server you wish to connect to
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
