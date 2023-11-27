package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"sync"
	"time"

	"logAgent/conf"
	"logAgent/etcd"
	"logAgent/kafka"
	"logAgent/taillog"
	"logAgent/utils"
)

var (
	// 全局结构体指针变量
	c = new(conf.AppConf) // 注意用new()初始化,否则会空指针
)

func main() {
	// 高级版解析配置
	// c := new(config)
	err := ini.MapTo(c, "./conf/config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	// 1.初始化kafka连接
	// 并起一个goroutine循环从通道取消息发往kafka
	err = kafka.Init(c.KafkaConf.Host, c.KafkaConf.Size)
	if err != nil {
		fmt.Printf("初始化 kafka 失败,err:%v\n", err)
	}
	fmt.Printf("初始化 Kafka 成功！\n")

	// 2.1 初始化etcd连接
	// 整型变量需要转换成时间变量
	_ = etcd.Init(c.EtcdConf.Host, time.Duration(c.EtcdConf.Timeout)*time.Second)
	if err != nil {
		fmt.Printf("初始化 etcd 失败,err:%v\n", err)
	}
	fmt.Printf("初始化 etcd 成功！\n")

	// 替换配置文件中的占位符
	addr := utils.GetLocalIP()
	logCollectionKey := fmt.Sprintf(c.EtcdConf.Key, addr)
	fmt.Println(logCollectionKey)

	// 2.2 从etcd中获取日志收集项的配置信息
	logEntryCollection := etcd.GetConf(logCollectionKey)
	// 遍历切片,打印需要收集的日志条目
	for idx, v := range logEntryCollection {
		fmt.Printf("索引: [%d], 值: %#v\n", idx, v)
	}

	// 3.根据获取的配置收集日志，创建多个TailTask
	// 每个TailTask起一个goroutine条用tailtask.run()异步往kafka发送消息
	// 启动一个goroutine调用tskMgr.run()异步从 newConfChan 获取新配置
	taillog.Init(logEntryCollection)

	// 4.etcd派一个哨兵监测logEntryCollection的变化
	// 通过taillog暴露的函数获取 newConfChan
	newConfCh := taillog.NewConfChan()
	var wg sync.WaitGroup
	wg.Add(1)
	// 起一个goroutine异步监测logEntryCollection的变化,如果配置有变化把配置发送给 newConfChan
	go etcd.WatchNewConf(logCollectionKey, newConfCh)
	wg.Wait()
}
