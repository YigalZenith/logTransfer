package conf

type EsConf struct {
	Address string `ini:"address"`
}

type KafkaConf struct {
	Address string `ini:"address"`
	Topic   string `ini:"topic"`
}

type AppConf struct {
	EsConf    `ini:"es"`
	KafkaConf `ini:"kafka"`
}
