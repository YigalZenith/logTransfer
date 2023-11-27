package conf

type KafkaConf struct {
	Host string `ini:"host"`
	Size int    `int:"maxChanSize"`
}

type EtcdConf struct {
	Host    string `ini:"host"`
	Timeout int    `ini:"timeout"`
	Key     string `ini:"log_collection"`
}

type AppConf struct {
	KafkaConf `ini:"kafka"` // 嵌套的结构体不能叫kafka 否则和导入的kafka包名冲突
	EtcdConf  `ini:"etcd"`
}
