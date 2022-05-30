package locate

import (
	"encoding/json"
	"oss/src/lib/rabbitmq"
	"oss/src/lib/rs"
	"oss/src/lib/types"
	"time"
)

func Locate(name string) (locateInfo map[int]string) {
	// rabbitmq.New创建一个rabbitmq.RabbitMQ结构体
	q := rabbitmq.New()
	q.Publish("dataServers", name)
	c := q.Consume()

	// 1秒后关闭mq
	go func() {
		time.Sleep(time.Second)
		q.Close()
	}()

	// map集合locateInfo
	locateInfo = make(map[int]string)

	// rs.ALL_SHARDS 常数6
	for i := 0; i < rs.ALL_SHARDS; i++ {
		// 从队列中读取消息
		msg := <-c
		if len(msg.Body) == 0 {
			return
		}
		// 创建一个包含地址和id的LocateMessage结构体 -- info
		var info types.LocateMessage

		//反序列化
		json.Unmarshal(msg.Body, &info)
		// map {"id","addr"}
		locateInfo[info.Id] = info.Addr
	}
	// 返回map
	return
}

func Exist(name string) bool {
	// 判断返回消息是否大于4 大于返回true，则存在
	return len(Locate(name)) >= rs.DATA_SHARDS
}
