package taillog

import (
	"fmt"
	"time"

	"logAgent/etcd"
)

var tskMgr *tailTaskMgr

// TailTask管理者
type tailTaskMgr struct {
	logEntryCollection []*etcd.LogEntry      // 上一版日志收集配置的集合
	tskMap             map[string]*TailTask  // 保存TailTask对象,后续根据对象进行热更新和启停相应的任务
	newConfChan        chan []*etcd.LogEntry // 新的日志配置集合
}

func Init(logEntryCollection []*etcd.LogEntry) {
	tskMgr = &tailTaskMgr{
		logEntryCollection: logEntryCollection, // 把当前的日志配置集合保存起来
		newConfChan:        make(chan []*etcd.LogEntry),
		tskMap:             make(map[string]*TailTask, 16),
	}

	// 循环配置集合的切片,每一个日志条目运行一个TailTask任务往kafka发送消息
	for _, logEntry := range logEntryCollection {
		// 拿到每个TailTask的对象
		tailTask := NewTailTask(logEntry.Path, logEntry.Topic)
		// 已每个日志收集条目的path+topic作为key
		mk := fmt.Sprintf("%s_%s", logEntry.Path, logEntry.Topic)
		// 把每个TailTask和它对应的key存入map中(即当前运行的所有日志收集任务的记录)
		tskMgr.tskMap[mk] = tailTask
	}

	go tskMgr.run()
}

// NewConfChan 暴露一个函数,让etcd.WatchNewConf()往通道发送新配置
func NewConfChan() chan<- []*etcd.LogEntry {
	return tskMgr.newConfChan
}

// 监听自己的newConfChan,有了新的配置过来之后做对应的处理
func (t *tailTaskMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan:

			fmt.Printf("配置发生了变化，%v\n", newConf)

			// 1.遍历新配置,如果在t.tskMap不存在，代表是新增或修改的配置
			for _, v := range newConf {
				mk := fmt.Sprintf("%s_%s", v.Path, v.Topic)
				if _, ok := t.tskMap[mk]; !ok {
					fmt.Printf("新增配置: %v\n", v)

					// 启动新的tailtask进行日志收集
					newTailTask := NewTailTask(v.Path, v.Topic)
					// 并把新的tailtask和其对应的Key存入map
					t.tskMap[mk] = newTailTask
					fmt.Printf("上一版配置为: %v\n", t.logEntryCollection)
				}
			}

			// // 2.遍历上一版配置，如果在新配置中不存在,代表是要删除的配置
			// for _, cur := range t.logEntryCollection {
			// 	flag := true
			// 	for _, new := range newConf {
			// 		if cur.Path == new.Path && cur.Topic == new.Topic {
			// 			flag = false
			// 			continue
			// 		}
			// 	}
			// 	if flag {
			// 		mk := fmt.Sprintf("%s_%s", cur.Path, cur.Topic)
			// 		// 停掉老的tailtask
			// 		t.tskMap[mk].cancel()
			// 		// 剔除老的tailtask的map映射
			// 		delete(t.tskMap, mk)
			// 		fmt.Printf("删除配置: %v\n", mk)
			// 		fmt.Printf("上一版配置为: %v\n", t.logEntryCollection)
			// 	}
			// }

			// 2.遍历上一版配置，如果在新配置中不存在,代表是要删除的配置
			for k, v := range t.tskMap {
				isDelete := true
				// 上一版配置每个k和新配置组合的mk进行比较,如果k在新配置中不存在,则需要剔除
				for _, new := range newConf {
					mk := fmt.Sprintf("%s_%s", new.Path, new.Topic)
					if k == mk {
						isDelete = false
					}
				}
				if isDelete {
					// 停掉老的tailtask
					v.cancel()
					// 剔除老的tailtask的map映射
					delete(t.tskMap, k)
					fmt.Printf("删除配置: %v\n", k)
					fmt.Printf("上一版配置为: %v\n", t.logEntryCollection)
				}
			}

			// 3.处理完新增或删除后
			t.logEntryCollection = newConf
			fmt.Printf("当前配置: %v\n", t.logEntryCollection)

		default:
			time.Sleep(time.Second)
		}
	}
}
