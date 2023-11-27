package kafka

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
)

var client sarama.SyncProducer

// 存放日志消息的通道
var logDataChan chan *logData

// 保存日志消息的结构体
type logData struct {
	msg   string
	topic string
}

// 初始化kafka连接,起一个goroutine循环从msg chan获取消息
func Init(addr string, maxChanSize int) (err error) {
	// 设置生产者的分区选择和ACK确认机制
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	client, err = sarama.NewSyncProducer([]string{addr}, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}

	// 初始化存放日志消息的通道
	logDataChan = make(chan *logData, maxChanSize)

	// 后台goroutine循环消费通道的消息,发送到kafka
	go chanToKafka()
	return
}

// 暴露一个函数让taillog调用,往日志消息通道中异步发送消息
func MsgToChan(topic, msg string) {
	logData := &logData{
		msg:   msg,
		topic: topic,
	}
	logDataChan <- logData
}

// 循环消费通道的消息,发送到kafka
func chanToKafka() {
	for {
		select {
		case v := <-logDataChan:
			// 构造一个消息
			msg := &sarama.ProducerMessage{}
			msg.Topic = v.topic
			msg.Value = sarama.StringEncoder(v.msg)

			// 发送消息
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Printf("pid:%v offset:%v\n", pid, offset)
		default:
			time.Sleep(time.Second)
		}
	}
}
