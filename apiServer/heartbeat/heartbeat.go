package heartbeat

import (
	"encoding/json"
	"fmt"
	"os"
	"oss/src/lib/dingTalk"
	"oss/src/lib/myLog"
	"oss/src/lib/rabbitmq"
	"oss/src/lib/redis"
	"strconv"
	"time"
)

// 类型为map，键string，值为time，用于缓存所有的数据服务节点
//var dataServers = make(map[string]time.Time)

//type dataServers struct {
//	addr string
//	time time.Time
//}
const SET = 0
const GET = 1
const DELETE = 2
const INCR = 3

// 互斥锁 ：go的多个纤程能够同时读map，但不支持同时写或同时既读又写，互斥锁保护map的并发读写
// 加锁后无论读写只允许一个纤程操作map

// ListenHeartbeat 接口服务对数据服务进行心跳检测
func ListenHeartbeat() {
	// rabbitmq.New创建一个rabbitmq.RabbitMQ结构体
	q := rabbitmq.New()

	// 延迟关闭mq
	defer q.Close()

	// 绑定exchange交换机：dataServers
	q.Bind("apiServers")

	// Consume方法返回一个go的通道
	c := q.Consume()

	// 启动一个纤程 执行removeExpiredDataServer
	go removeExpiredDataServer()
	go sendDindTalkReport()
	var dataServers = make(map[string]time.Time)
	// 遍历这个通道
	for msg := range c {
		// strconv.Unquote方法将msg.Body中字符串前后的双引号去除并返回
		dataServer, e := strconv.Unquote(string(msg.Body))

		// 如果报错
		if e != nil {
			// panic报错，打断程序运行，不影响defer延迟关闭mq
			panic(e)
		}

		//// 加锁
		//mutex.Lock()
		//// 监听到心跳，给服务节点心跳检测时间
		//dataServers[dataServer] = time.Now()
		//
		//// 释放锁
		//mutex.Unlock()

		// 监听到心跳，给服务节点心跳检测时间
		dataServers[dataServer] = time.Now()
		//改用redis记录
		marshal, e := json.Marshal(dataServers)
		redis.RedisSet("dataServer", marshal)
		redis.RedisSet("apiDataServer", marshal)
	}
}

func removeExpiredDataServer() {
	// 5秒循环一次
	for {
		time.Sleep(5 * time.Second)
		dataServers := redis.RedisGet("dataServer")
		serversMap := make(map[string]time.Time)
		err := json.Unmarshal([]byte(dataServers), &serversMap)
		if err != nil {
			myLog.Error.Println(err)
			//log.Print(err)
		}
		// 遍历所有的数据服务节点dataServers
		for s, t := range serversMap {
			// 如果10s没有收到心跳消息的数据服务节点
			if t.Add(10 * time.Second).Before(time.Now()) {
				// 删除节点
				delete(serversMap, s)
				myLog.Warn.Println(fmt.Sprintf("检测到%s节点下线", s))
				marshal, _ := json.Marshal(serversMap)
				redis.RedisSet("dataServer", marshal)
			}
		}
	}
}

// 向钉钉机器人发送监控告警
func sendDindTalkReport() {
	// 30秒循环一次
	for {
		time.Sleep(30 * time.Second)
		dataServers := redis.RedisGet("apiDataServer")
		serversMap := make(map[string]time.Time)
		err := json.Unmarshal([]byte(dataServers), &serversMap)
		if err != nil {
			myLog.Error.Println(err)
			//log.Print(err)
		}
		// 遍历所有的数据服务节点apiDataServer
		for s, t := range serversMap {
			// 如果10s没有收到心跳消息的数据服务节点
			if t.Add(10 * time.Second).Before(time.Now()) {
				myLog.Info.Println("钉钉群聊机器人发送告警")
				// 向钉钉机器人发送消息
				text := fmt.Sprintf("### 接口服务心跳检测发现异常\n\n> 发起报告节点：%s\n>\n> 报告时间：%s\n>\n> 状态：紧急（30s将重复一次提醒）\n\n**数据服务故障主机IP**：%s\n\n**最后一次心跳检测时间**：%s\n\n请相关人员上服务器检查！！！", os.Getenv("LISTEN_ADDRESS"), time.Now().Format("2006-01-02 15:04:05"), s, t.Format("2006-01-02 15:04:05"))
				go dingTalk.RobotSend(text)
			}
		}
	}
}

// GetDataServers 遍历所有保持心跳检测的数据服务节点
func GetDataServers() []string {
	// 加互斥锁
	//mutex.Lock()
	// 延迟释放锁
	//defer mutex.Unlock()

	ds := make([]string, 0)

	dataServers := redis.RedisGet("dataServer")
	maps := make(map[string]time.Time)
	err := json.Unmarshal([]byte(dataServers), &maps)
	if err != nil {
		myLog.Error.Println(err)
	}

	// 遍历所有的数据服务节点dataServers
	for s, _ := range maps {
		ds = append(ds, s)
	}

	// 返回所有数据节点
	return ds
}

func GetDataServersMap() map[string]time.Time {
	// 加互斥锁
	//mutex.Lock()
	// 延迟释放锁
	//defer mutex.Unlock()
	dataServers := redis.RedisGet("apiDataServer")
	maps := make(map[string]time.Time)
	err := json.Unmarshal([]byte(dataServers), &maps)
	if err != nil {
		myLog.Error.Println(err)
	}
	// 返回所有数据节点
	return maps
}
