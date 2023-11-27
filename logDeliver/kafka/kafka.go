package kafka

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"

	"logdeliver/es"
)

func Init(addr, topic string) {
	consumer, err := sarama.NewConsumer([]string{addr}, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		return
	}

	partitionList, err := consumer.Partitions(topic) // 根据topic取到所有的分区
	if err != nil {
		fmt.Printf("fail to get list of partition:err%v\n", err)
		return
	}
	fmt.Println("connect to kafka success.")
	fmt.Println(partitionList)

	for partition := range partitionList { // 遍历所有的分区
		// 针对每个分区创建一个对应的分区消费者
		pc, err := consumer.ConsumePartition(topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("failed to start consumer for partition %d,err:%v\n", partition, err)
			return
		}

		// 异步从每个分区消费信息
		go func(sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				//fmt.Printf("Partition:%d Offset:%d Key:%v Value:%s\n", msg.Partition, msg.Offset, msg.Key, msg.Value)

				// 手动构造json发送给ES
				// data := `{"time":20210913,"content":"haha"}`

				// 构造map发送给ES
				//var m = map[string]interface{}{}
				//m["date"] = time.Now().Format("2006/01/02 15:04:05")
				//m["content"] = string(msg.Value)

				// 构造结构体发送给ES
				// ld := &logData{time.Now().Format("2006/01/02 15:04:05"), string(msg.Value)}
				ld := es.LogData{
					Topic:   topic,
					Date:    time.Now().Format("2006/01/02 15:04:05"),
					Content: string(msg.Value),
				}

				// map或结构体可以直接发给ES(会自动转为json), 也可以手动转成json
				// 结构体字段名就是存入ES的字段名(区分大小写), 如果要使用小写的es字段,需要加tag(例: `json:"name"`)
				// es.SendToEs(topic, data) // 函数掉函数,代码执行效率低,改为通道
				es.SendToChan(&ld)
			}
		}(pc)
	}
}
