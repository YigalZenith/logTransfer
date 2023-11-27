package main

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 为logAgent包生产etcd键值对

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	fmt.Println("connect to etcd success")
	defer cli.Close()

	// put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 1.初始配置
	val := `[{"path":"/tmp/nginx.log","topic":"web_log"},{"path":"/tmp/redis.log","topic":"redis_log"}]`
	// 2.修改配置
	//val := `[{"path":"/tmp/nginx.log","topic":"nginx_log"},{"path":"/tmp/redis.log","topic":"redis_log"}]`
	// 3.新增配置
	//val := `[{"path":"/tmp/nginx.log","topic":"nginx_log"},{"path":"/tmp/redis.log","topic":"redis_log"},{"path":"/tmp/mysql.log","topic":"mysql_log"}]`
	_, err = cli.Put(ctx, "/logagent/127.0.0.1/log_collection_key", val)
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed, err:%v\n", err)
		return
	}
}
