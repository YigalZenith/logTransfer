package es

import (
	"context"
	"fmt"
	"strings"

	"github.com/olivere/elastic/v7"
)

var (
	client *elastic.Client
	ch     chan *LogData
)

type LogData struct {
	Topic   string `json:"topic"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

func Init(addr string) (err error) {
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}
	client, err = elastic.NewClient(elastic.SetURL(addr))
	if err != nil {
		fmt.Printf("连接es失败,err:%v", err)
		return
	}
	fmt.Println("connect to es success")
	ch = make(chan *LogData, 10000)
	go sendToEs()
	return
}

// SendToEs 暴露给其他模块调用,函数版
// data: 可以是map或结构体(会备自动转为Json),也可以手动传json
// func SendToEs(idx string, data interface{}) {
// 	fmt.Printf("%#v\n", data)
// 	_, err := client.Index().Index(idx).BodyJson(data).Do(context.Background())
// 	if err != nil {
// 		fmt.Printf("es插入数据失败，err:%v\n", err)
// 		return
// 	}
// 	fmt.Printf("发送数据 %v 到ES索引 %s 中\n", data, idx)
// }

// SendToChan 暴露给其他模块调用,chan版
func SendToChan(ld *LogData) {
	ch <- ld
}

func sendToEs() {
	for v := range ch {
		fmt.Printf("%#v\n", v)
		_, err := client.Index().Index(v.Topic).BodyJson(v).Do(context.Background())
		if err != nil {
			fmt.Printf("es插入数据失败，err:%v\n", err)
			return
		}
		fmt.Printf("发送数据 %v 到ES索引 %s 中\n", v, v.Topic)
	}
}
