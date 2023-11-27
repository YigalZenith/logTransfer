package taillog

import (
	"context"
	"fmt"
	"time"

	"github.com/hpcloud/tail"

	"logAgent/kafka"
)

// *tail.Tail的管理者
// 一个TailTask对象代表一个收集日志的任务
type TailTask struct {
	path    string
	topic   string
	tailObj *tail.Tail
	ctx     context.Context
	cancel  context.CancelFunc
}

// 初始化TailTask对象
func NewTailTask(path, topic string) *TailTask {

	ctx, cancel := context.WithCancel(context.Background())

	tailTask := &TailTask{
		path:   path,
		topic:  topic,
		ctx:    ctx,
		cancel: cancel,
	}

	tailTask.init() // 根据路径去打开对应的日志

	return tailTask
}

func (t *TailTask) init() {
	config := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的哪个地方开始读
		MustExist: false,                                // 文件不存在不报错
		Poll:      true,                                 // Poll for file changes instead of using inotify
	}

	var err error
	t.tailObj, err = tail.TailFile(t.path, config)
	if err != nil {
		fmt.Println("tail file failed:,err:", err)
	}

	fmt.Printf("启动了新的tailtask: %v\n", t.path)
	go t.run()
}

// 每个tailtask独享循环从自己的tailObj通道取日志,借助通道异步发送到kafka
func (t *TailTask) run() {
	for {
		select {
		case <-t.ctx.Done():
			fmt.Printf("结束了老的tailtask: %v\n", t.path)
			return
		case line := <-t.tailObj.Lines:
			// kafka.SendMsg(t.topic, line.Text)   // run函数调用SendMsg函数,同步效率低,需要通过chan改成异步
			kafka.MsgToChan(t.topic, line.Text)
		default:
			time.Sleep(time.Second)
		}
	}
}
