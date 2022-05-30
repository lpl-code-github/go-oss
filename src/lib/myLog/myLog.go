package myLog

import (
	"fmt"
	"github.com/nxadm/tail"
	"log"
	"os"
	"oss/src/lib/es"
	"oss/src/lib/rabbitmq"
	"strconv"
	"strings"
	"time"
)

var (
	Trace *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

// 初始化方法自定义log
func init() {
	osName, _ := os.Hostname()

	file, err := os.OpenFile(fmt.Sprintf("%s%s.log", os.Getenv("LOG_DIRECTORY"), time.Now().Format("2006-01-02")), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error myLog file: ", err)
	}

	Trace = log.New(file, os.Getenv("LISTEN_ADDRESS")+"-"+osName+" [TRACE] ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(file, os.Getenv("LISTEN_ADDRESS")+"-"+osName+" [INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(file, os.Getenv("LISTEN_ADDRESS")+"-"+osName+" [WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(file, os.Getenv("LISTEN_ADDRESS")+"-"+osName+" [ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// ListenLogExchange 监听log交换机获取日志
func ListenLogExchange() {
	// rabbitmq.New创建一个rabbitmq.RabbitMQ结构体
	q := rabbitmq.New()

	// 延迟关闭mq
	defer q.Close()

	// 绑定exchange交换机：dataServers
	q.Bind("log")

	// Consume方法返回一个go的通道
	c := q.Consume()

	// 遍历这个通道
	for msg := range c {
		// strconv.Unquote方法将msg.Body中字符串前后的双引号去除并返回
		msgInfo, e := strconv.Unquote(string(msg.Body))
		// 如果报错
		if e != nil {
			// panic报错，打断程序运行，不影响defer延迟关闭mq
			panic(e)
		}
		//log.Println(msgInfo)
		// 向es推送log
		go es.PutLog(msgInfo)
	}
}

// ReadLog 实时读取日志文件并推送到RabbitMQ
func ReadLog(times string) {
	fileName := fmt.Sprintf("%s%s.log", os.Getenv("LOG_DIRECTORY"), times)
	config := tail.Config{
		ReOpen:    true,                                 // 打开文件
		Follow:    true,                                 // 文件切割自动重新打开
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // Location读取文件的位置, Whence更加系统选择参数
		MustExist: false,                                // 允许日志文件不存在
		Poll:      true,                                 // 轮询
	}
	// 打开文件读取日志
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		Error.Println("tail file failed, err:", err)
		return
	}
	// 开始读取数据
	var (
		msg *tail.Line
		ok  bool
	)
	for {
		msg, ok = <-tails.Lines
		if !ok {
			Error.Printf("tail file close reopen, filename:%s\n", tails.Filename)
			time.Sleep(time.Second) // 读取出错停止一秒
			continue
		}
		//fmt.Println(msg.Text)
		logMsg := analysisLog(msg.Text)
		// 将日志推送到mq
		go publishLog("log", string(logMsg))
	}
}

// 解析某条log
func analysisLog(logContent string) string {
	//l := &Log{}
	split := strings.Split(logContent, " ")
	osName := split[0]
	level := strings.Trim(split[1], "[]")
	date := strings.Replace(split[2], "/", "-", -1)
	time := split[3]
	content := strings.Split(logContent, fmt.Sprintf("%s %s %s %s ", osName, split[1], split[2], time))[1]
	doc := fmt.Sprintf(`{"osName":"%s","level":"%s","dateTime":%d,"content":%s}`,
		osName, level, dateFormat(fmt.Sprintf("%s %s", date, time)), strconv.Quote(content))
	return doc
}

// 向Rabbit MQ推送日志
func publishLog(exchange, msg string) {
	var q *rabbitmq.RabbitMQ
	q = rabbitmq.New()
	defer q.Close()

	// Publish方法 发送一条消息
	err := q.Publish(exchange, msg)
	if err != nil { //如果发送发生异常 尝试重新发送 因为NewMQ会选取新的节点
		publishLog(exchange, msg)
	}
}

// 字符串时间转时间戳工具
func dateFormat(dataTime string) int64 {
	loc, _ := time.LoadLocation("Local")
	formatTime, err := time.ParseInLocation("2006-01-02 15:04:05", dataTime, loc)

	if err != nil {
		Error.Printf("filed datetime format")
	}
	return formatTime.Unix()
}
