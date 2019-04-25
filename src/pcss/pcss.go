// 1.建立Star
// 2.侦听UDP端口
// 3.UDP端口用于代理
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
	Version           = "0.1.0-190424"    // 当前版本号
	LogFile           = "pcs.log"         // Log文件名
	ConfigFile        = "pcs.config.json" // 配置文件名
	DefaultID         = "auto"            // 默认ID（或生成方式）
	DefaultServerPort = 3478              // 默认本地服务器端口
)

var (
	logger     *log.Logger // 全局Logger
	config     Config      // 全局配置信息
	ID         string      // 服务器ID（一机一个）
	ServerPort int         // 本地服务器端口
)

func init() { // 初始化
	os.Remove(LogFile) // 删除记录文件（如果有）
	// 指定记录文件
	logFile, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Println(err)
	}
	// defer logFile.Close()
	// 记录文件和控制台双通
	mw := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(mw, "", log.LstdFlags)
}
func main() {
	logger.Println("[HALO]", "Point Cloud System Server", "[Version", Version+"]")
	logger.Println("[HALO]", "Welcome to PCS Server!")
	loadConfig()
	logger.Println("[INFO]", "Current ID:", ID)
	startServer()
}
func loadConfig() {
	// 处理配置文件
	config = Config{
		ID:         DefaultID,
		ServerPort: DefaultServerPort,
	}
	logger.Println("[INFO]", "Searching Config File...")
	if _, err := os.Stat(ConfigFile); err == nil {
		// 配置文件存在
		logger.Println("[INFO]", "Loading Config File...")
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
		logger.Println("[INFO]", "Creating Config File...")
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
}
func startServer() {
	ServerPort = config.ServerPort
	for {
		logger.Println("[INFO]", "Trying to set Server Port on", ServerPort, "...")
		l, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: ServerPort})
		if err != nil {
			logger.Println("[ERRO]", "The Server Port is Now in use!")
			ServerPort += 1
			continue
		} else {
			defer l.Close()
			handleConnection(l)
		}
	}
}
func handleConnection(l *net.UDPConn) {
	// reader := bufio.NewReader(l)
	// writer := bufio.NewWriter(l)
	// readwriter := bufio.NewReadWriter(reader, writer)
	logger.Println("[INFO]", "Listening...")
	var data []byte
	for {
		// 等待接入
		_, remoteAddr, err := l.ReadFromUDP(data)
		if err != nil {
			logger.Fatal(err)
		}
		l.WriteToUDP(data, remoteAddr)
		// if m["CMD"] == "close" {
		// 	logger.Println("[INFO]", "Star-Point Recieved Close Signal!")
		// 	break
		// } else {
		// 	logger.Println("[INFO]", "Recieved Request from Star:", m["CMD"])
		// 	if m["CMD"] == "signup" {
		// 		logger.Println("[INFO]", "Register Star ID:", m["ID"])
		// 	} else if m["CMD"] == "login" {
		// 		logger.Println("[INFO]", "Login User:", m["Username"])
		// 	}
		// }
	}
}

// func startServer() {
// 	ServerPort = config.ServerPort
// 	for {
// 		logger.Println("[INFO]", "尝试在", ServerPort, "端口搭建核心服务器...")
// 		l, err := net.Listen("tcp", ":"+strconv.Itoa(ServerPort))
// 		if err != nil {
// 			logger.Println("[ERRO]", ServerPort, "核心服务器端口被占用！")
// 			ServerPort += 1
// 			continue
// 		} else {
// 			defer l.Close()
// 			for {
// 				// 等待接入
// 				logger.Println("[INFO]", "核心服务器侦听中...")
// 				conn, err := l.Accept()
// 				if err != nil {
// 					logger.Fatal(err)
// 				}
// 				logger.Println("[FINE]", "建立Star-Point连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
// 				// 在新的Go程里处理会话
// 				// 循环返回到等待新接入，就可以用协程处理接入
// 				go handleConnection(conn)
// 			}
// 		}
// 	}
// }

// func handleConnection(conn net.Conn) {
// 	reader := bufio.NewReader(conn)
// 	writer := bufio.NewWriter(conn)
// 	readwriter := bufio.NewReadWriter(reader, writer)
// 	for {
// 		m := ReadMap(readwriter)
// 		if m["CMD"] == "close" {
// 			logger.Println("[INFO]", "Star-Point连接接收到关闭信号！")
// 			break
// 		} else {
// 			logger.Println("[INFO]", "接收到Star请求：", m["CMD"])
// 			if m["CMD"] == "signup" {
// 				WriteMap(readwriter, m)
// 			} else if m["CMD"] == "login" {
// 				WriteMap(readwriter, m)
// 			}
// 		}
// 	}
// 	m := map[string]interface{}{
// 		"CMD": "close",
// 	}
// 	WriteMap(readwriter, m)
// 	conn.Close()
// 	logger.Println("[INFO]", "Star-Point连接受控关闭！")
// }
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
		logger.Fatal(err)
	}
	msg := strings.Split(raw_msg, "\n")
	return msg[0]
}
func WriteString(readWriter *bufio.ReadWriter, str string) {
	readWriter.WriteString(str + "\n")
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
	Mode       []string // 支持的服务模式："STUN":去中心化，消耗服务器资源低；"TURN":中心化，消耗服务器资源高；"ICE":智能模式
	ID         string   // 本机ID
	ServerPort int      // 本机ID
	ProxyPort  int      // Port for Data Transfer
}
type PointInfo struct {
	ID          string
	GlobalAddr  string
	GlobalPort  int
	LocalAddr   string
	LocalPort   int
	NatType     string // NAT类型：完全锥形、限制锥形、端口限制锥形、对称
	NeighborID1 string
	NeighborID2 string
	NeighborID3 string
}
type CloudInfo struct {
	ID     string
	Points []PointInfo
}
type StarInfo struct {
	ID     string
	Addr   string
	Port   string
	Clouds []CloudInfo
}
