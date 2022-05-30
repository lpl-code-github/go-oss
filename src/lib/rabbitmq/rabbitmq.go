package rabbitmq

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"math/rand"
	"os"
	"strings"
	"time"
)

type RabbitMQ struct {
	channel  *amqp.Channel
	conn     *amqp.Connection
	Name     string
	exchange string
}

var (
	available           []string // 可用节点
	rabbitmqAddrsString = os.Getenv("RABBITMQ_SERVER")
	rabbitmqAddr        = strings.Split(rabbitmqAddrsString, ",")
)

func init() {
	available = rabbitmqAddr
	go apendAddr()
}

// 遍历所有节点
func apendAddr() {
	for {
		result := make([]string, 0)
		// 遍历节点
		for i, addr := range rabbitmqAddr {
			_, e := amqp.Dial(addr) // 尝试连接
			if e != nil {           // 如果不能连接 下一次循环
				if i+1 == len(rabbitmqAddr) { // 如果最后一个节点也不可用 记录次数
					panic(e)
				}
				continue
			}
			result = append(result, addr)
		}
		available = result
		// 延时5秒
		time.Sleep(5 * time.Second)
	}
}

// 选取一个节点
func choose(availableAddrs []string) string {
	// 从可用节点中随机选取一个节点
	addr := rand.Intn(len(available))
	_, e := amqp.Dial(available[addr])
	if e != nil {
		return choose(availableAddrs)
	}
	return available[addr]
}

func New() *RabbitMQ {
	conn, e := amqp.Dial(choose(available)) //建立连接
	if e != nil {
		return New() // 如果连接失败，重新选举一个新的节点
	}

	ch, e := conn.Channel()
	if e != nil {
		panic(e)
	}
	q, e := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if e != nil {
		panic(e)
	}

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.conn = conn
	mq.Name = q.Name
	return mq
}

func (q *RabbitMQ) Bind(exchange string) {
	e := q.channel.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil)
	if e != nil {
		panic(e)
	}
	q.exchange = exchange
}

func (q *RabbitMQ) Send(queue string, body interface{}) {
	str, e := json.Marshal(body)
	if e != nil {
		panic(e)
	}
	e = q.channel.Publish("",
		queue,
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	if e != nil {
		panic(e)
	}
}

func (q *RabbitMQ) Publish(exchange string, body interface{}) error {
	str, e := json.Marshal(body)
	if e != nil {
		panic(e)
	}
	err := q.channel.Publish(exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	return err
}

func (q *RabbitMQ) Consume() <-chan amqp.Delivery {
	c, e := q.channel.Consume(q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if e != nil {
		panic(e)
	}
	return c
}

func (q *RabbitMQ) Close() {
	q.channel.Close()
	q.conn.Close()
}
