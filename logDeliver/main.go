package main

import (
	"fmt"

	"logdeliver/conf"
	"logdeliver/es"
	"logdeliver/kafka"

	"gopkg.in/ini.v1"
)

func main() {
	// 1.加载配置文件
	var cfg conf.AppConf
	err := ini.MapTo(&cfg, "./conf/config.ini")
	if err != nil {
		fmt.Printf("加载配置文件失败,err:%v", err)
	}
	fmt.Println(cfg)

	// 2.初始化ES
	// 2.1 对外提供一个往ES写数据的函数
	err = es.Init(cfg.EsConf.Address)
	if err != nil {
		fmt.Printf("连接ES失败,err:%v", err)
		return
	}

	// 3.初始化kafka
	// 3.1 创建kafka消费者
	// 3.2 每个分区的消费者分别取出数据 通过ES提供的的函数往ES写数据
	kafka.Init(cfg.KafkaConf.Address, cfg.KafkaConf.Topic)
	if err != nil {
		fmt.Printf("连接kafka失败,err:%v", err)
		return
	}

	select {}
}
