// 1.连接Point
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
	Version      = "0.1.0-190424"
	LogFile      = "pct.log"
	ConfigFile   = "pct.config.json"
	ProxyAddress = "localhost"
	ProxyPort    = 1996
)

var (
	logger *log.Logger
	config Config
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
	logger.Println("[HALO]", "Point Cloud System Tester", "[版本", Version+"]")
	logger.Println("[HALO]", "欢迎使用点云测试程序！")
	// 处理配置文件
	config = Config{
		ProxyAddress: ProxyAddress,
		ProxyPort:    ProxyPort,
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
	startTester()
}
func startTester() {
	logger.Println("[INFO]", "正在尝试连接点云客户端", config.ProxyAddress+":"+strconv.Itoa(config.ProxyPort), "...")
	conn, err := net.Dial("tcp", config.ProxyAddress+":"+strconv.Itoa(config.ProxyPort))
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("[FINE]", "建立连接：", conn.LocalAddr(), "<==>", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readwriter := bufio.NewReadWriter(reader, writer)
	logger.Println("[TEST]", "开始测试！")
	logger.Println("[TEST]", "提交登录请求...")
	m := map[string]interface{}{
		"CMD":      "login",
		"Username": "abc",
		"Password": "123",
	}
	WriteMap(readwriter, m)
	logger.Println("[TEST]", "等待回应...")
	m = ReadMap(readwriter)
	logger.Println("[TEST]", "收到测试消息：", m)
	logger.Println("[TEST]", "发送关闭信号...")
	m = map[string]interface{}{
		"CMD": "close",
	}
	WriteMap(readwriter, m)
	logger.Println("[FINE]", "测试完毕！")
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
	raw_msg, _ := readWriter.ReadString('\n')
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
	ProxyAddress string // Address for Data Transfer
	ProxyPort    int    // Port for Data Transfer
	Username     string // Username For Star/Cloud Authentication
	Password     string // Password For Star/Cloud Authentication
}
