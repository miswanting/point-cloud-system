// 1.建立Point
// 2.侦听三个TCP端口和一个UDP端口
// 4.一个TCP端口用于与应用进行通信
// 5.两个TCP端口用于与相邻点进行通信
// 5.一个UDP端口用于多功能探针（与服务器通信/内网穿透）
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
	Version              = "0.1.0-190424"    // 当前版本号
	LogFile              = "pcc.log"         // Log文件名
	ConfigFile           = "pcc.config.json" // 配置文件名
	DefaultID            = "auto"            // 默认ID（或生成方式）
	DefaultServerAddress = "localhost"       // 默认服务器系统IP
	DefaultServerPort    = 3478              // 默认服务器系统端口
	DefaultClientPort    = 3479              // 默认本地系统端口
	DefaultProxyPort     = 3480              // 默认本地代理端口
)

var (
	logger       *log.Logger       // 全局Logger
	config       Config            // 全局配置信息
	ID           string            // 客户端ID（一机一个）
	SystemPort   int               // 系统端口
	ProxyPort    int               // 代理端口
	AppTCP       *bufio.ReadWriter // TCP连接App
	NeighborTCP1 *bufio.ReadWriter // TCP连接Neighbor1
	NeighborTCP2 *bufio.ReadWriter // TCP连接Neighbor2
	UDProbe      *bufio.ReadWriter // UDP探针
)

func init() { // 初始化
	os.Remove(LogFile) // 删除记录文件（如果有）
	// 设置记录文件
	logFile, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Println(err)
	}
	// defer logFile.Close()
	// 记录文件输出和控制台输出双通
	mw := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(mw, "", log.LstdFlags)
}
func main() {
	logger.Println("[HALO]", "Point Cloud System Client", "[版本", Version+"]")
	logger.Println("[HALO]", "欢迎使用点云客户端！")
	loadConfig()
	logger.Println("[INFO]", "当前ID:", ID)
	// 尝试连接Star
	go connectToServer()
	// 启动代理服务器
	go startProxyServer()
	startClient()
}
func loadConfig() {
	// 处理配置文件
	config = Config{
		ID:            DefaultID,
		ClientPort:    DefaultClientPort,
		ProxyPort:     DefaultProxyPort,
		ServerAddress: DefaultServerAddress,
		ServerPort:    DefaultServerPort,
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
}
func startClient() {
	ClientPort := config.ClientPort
	for {
		logger.Println("[INFO]", "尝试在端口", ClientPort, "搭建控制台...")
		l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: ClientPort})
		if err != nil {
			logger.Println("[ERRO]", "控制台端口被占用！")
			ClientPort += 1
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
func startProxyServer() {
	ProxyPort = config.ProxyPort
	for {
		logger.Println("[INFO]", "尝试在端口", ProxyPort, "搭建代理服务器...")
		l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4zero, Port: ProxyPort})
		if err != nil {
			logger.Println("[ERRO]", "代理服务器端口被占用！")
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
func connectToServer() {
	ServerAddress := config.ServerAddress
	ServerPort := config.ServerPort
	logger.Println("[INFO]", "正在尝试连接", ServerAddress+":"+strconv.Itoa(ServerPort), "...")
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(ServerAddress), Port: ServerPort})
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("[FINE]", "建立连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	UDProbe = bufio.NewReadWriter(reader, writer)
	m := map[string]interface{}{
		"CMD": "signup",
		"ID":  ID,
	}
	logger.Println(m)
	WriteMap(UDProbe, m)
	m = ReadMap(UDProbe)
	logger.Println(m)
}

// func OLD_connectToStar() {
// 	logger.Println("[INFO]", "正在尝试连接点云服务器", config.StarAddress+":"+strconv.Itoa(config.StarPort), "...")
// 	conn, err := net.Dial("tcp", config.StarAddress+":"+strconv.Itoa(config.StarPort))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("[FINE]", "建立Point-Star连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
// 	reader := bufio.NewReader(conn)
// 	writer := bufio.NewWriter(conn)
// 	StarTCP = bufio.NewReadWriter(reader, writer)
// 	m := map[string]interface{}{
// 		"CMD": "signup",
// 		"ID":  ID,
// 	}
// 	WriteMap(StarTCP, m)
// 	for {
// 		m = ReadMap(StarTCP)
// 		if m["CMD"] == "close" {
// 			logger.Println("[INFO]", "Point-Star连接接收到关闭信号！")
// 			break
// 		} else {
// 			logger.Println("[INFO]", "接收到Star请求：", m["CMD"])
// 			if m["CMD"] == "login" {
// 				WriteMap(AppTCP, m)
// 			}
// 		}
// 	}
// 	m = map[string]interface{}{
// 		"CMD": "close",
// 	}
// 	WriteMap(StarTCP, m)
// 	conn.Close()
// 	logger.Println("[INFO]", "Point-Star连接受控关闭！")
// }
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
	AppTCP = bufio.NewReadWriter(reader, writer)
	for {
		m := ReadMap(AppTCP)
		if m["CMD"] == "close" {
			logger.Println("[INFO]", "Point-App连接接收到关闭信号！")
			break
		} else {
			logger.Println("[INFO]", "接收到App请求：", m["CMD"])
			if m["CMD"] == "login" {
				WriteMap(UDProbe, m)
			}
		}
	}
	m := map[string]interface{}{
		"CMD": "close",
	}
	WriteMap(AppTCP, m)
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
	Mode          string // General Mode
	ID            string // Point ID ("auto"/)
	Username      string // Username For Star/Cloud Authentication
	Password      string // Password For Star/Cloud Authentication
	ClientPort    int    // Port for PCS System Data
	ProxyPort     int    // Port for Data Transfer
	ServerAddress string // IP address for Point Cloud Server you wish to connect to
	ServerPort    int    // Port for Point Cloud Server you wish to connect to
}
type PointInfo struct { // 只知道Point的ID和相邻关系
	ID        string
	Neighbor1 string
	Neighbor2 string
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
