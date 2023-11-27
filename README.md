# 简介

- logAgent: 结合 goroutine 和 chan 实现并发收集日志, 然后发送到 kafka 中，并结合 etcd 的 watch 实现配置热更新
- logDeliver: 并发从 kafka 中消费数据, 然后输出到 ES 中

# 依赖的中间件

```
# 创建网络
docker network create app-tier --driver bridge

# 运行 etcd
docker run -d --name Etcd-server \
    --network app-tier \
    --publish 2379:2379 \
    --publish 2380:2380 \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_ADVERTISE_CLIENT_URLS=http://127.0.0.1:2379 \
    bitnami/etcd:latest

# 运行 kafka
docker run -d --name kafka-server \
    --hostname kafka-server \
    --network app-tier \
    -e KAFKA_CFG_NODE_ID=0 \
    -e KAFKA_CFG_PROCESS_ROLES=controller,broker \
    -e KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
    -e KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
    -e KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=0@kafka-server:9093 \
    -e KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER \
    -p 9092:9092 \
    -p 2181:2181 \
    bitnami/kafka:latest

# 运行 ES
docker run -d --name elasticsearch \
    --net app-tier \
    -p 9200:9200 \
    -p 9300:9300 \
    -e "discovery.type=single-node" \
    elasticsearch:7.17.15
```
