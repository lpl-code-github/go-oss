package heartbeat

import (
	"os"
	"oss/src/lib/rabbitmq"
	"time"
)

func StartHeartbeat() {
	// rabbitmq.New创建一个rabbitmq.RabbitMQ结构体
	var q *rabbitmq.RabbitMQ
	q = rabbitmq.New()
	defer q.Close()

	// 5秒循环一次
	for {
		// Publish方法 向mq的apiServers交换机发送一条消息，把本服务的节点监听地址发送出去
		err := q.Publish("apiServers", os.Getenv("LISTEN_ADDRESS"))
		if err != nil { //如果发送发生异常 重新连接 并跳到下一次循环
			q = rabbitmq.New()
			continue
		}
		time.Sleep(5 * time.Second)
	}
}
