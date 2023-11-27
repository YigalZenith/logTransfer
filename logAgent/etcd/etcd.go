package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var cli *clientv3.Client

// 需要收集的日志的配置信息
type LogEntry struct {
	Path  string `json:"path"`  // 日志存放的路径
	Topic string `json:"topic"` // 日志发往kafka的topic
}

func Init(addr string, tw time.Duration) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: tw,
	})
	if err != nil {
		// handle error!
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return err
	}
	return
}

func GetConf(logCollectionKey string) (logEntryCollection []*LogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	getResult, err := cli.Get(ctx, logCollectionKey)
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
	}
	for _, kv := range getResult.Kvs {
		// 把etcd中获取的条目，反序列化到 []*LogEntry
		err := json.Unmarshal(kv.Value, &logEntryCollection)
		if err != nil {
			fmt.Printf("unmarshal LogEntryConf failed,err:%v\n", err)
		}
	}
	return
}

// WatchNewConf 监听到配置变化后往 newConfChan 通道发送新的配置
func WatchNewConf(logCollectionKey string, newConfCh chan<- []*LogEntry) {
	watchChan := cli.Watch(context.Background(), logCollectionKey)
	for watchResponse := range watchChan {
		for _, evt := range watchResponse.Events {
			// fmt.Printf("Type:%#v Key:%s Value:%s \n", evt.Type, evt.Kv.Key, evt.Kv.Value)
			// 通知taillog.taskMgr做配置变更
			// 1.先判断操作的类型
			var newConf []*LogEntry

			// 如果是删除操作如何处理
			//if evt.Type != clientv3.EventTypeDelete {
			// 略
			//}

			err := json.Unmarshal(evt.Kv.Value, &newConf)
			if err != nil {
				fmt.Printf("etcd.WatchNewConf unmarshal failed, err:%v\n", err)
				continue
			}
			newConfCh <- newConf
		}
	}
}
