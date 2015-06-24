package node

import (
	"encoding/json"
	"github.com/henrylee2cn/pholcus/node/crawlpool"
	. "github.com/henrylee2cn/pholcus/node/net"
	"github.com/henrylee2cn/pholcus/node/spiderqueue"
	"github.com/henrylee2cn/pholcus/node/task"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"github.com/henrylee2cn/pholcus/runtime/status"
	"log"
	"strconv"
	"time"
)

type Node struct {
	*Network
	// 节点间传递的任务的存储库
	tasks *task.TaskJar
	// 当前任务的蜘蛛队列
	Spiders spiderqueue.SpiderQueue
	// 爬行动作的回收池
	Crawls crawlpool.CrawlPool
	// 节点状态
	Status int
}

func newPholcus() *Node {
	return &Node{
		Network: &Network{
			RunMode: cache.Task.RunMode,
			Port:    ":" + strconv.Itoa(cache.Task.Port),
			Master:  cache.Task.Master,
			Conns:   map[string]*Conn{},
		},
		tasks:   task.NewTaskJar(),
		Spiders: spiderqueue.New(),
		Crawls:  crawlpool.New(),
		Status:  status.RUN,
	}
}

// 声明实例
var Pholcus *Node = nil

// 运行节点
func PholcusRun() {
	if Pholcus != nil {
		return
	}
	Pholcus = newPholcus()
	switch Pholcus.GetRunMode() {
	case status.SERVER:
		if Pholcus.checkPort() {
			log.Printf("                                                                                                          ！！当前运行模式为：[ 服务器 ] 模式！！")
			go Pholcus.Server()
		}

	case status.CLIENT:
		if Pholcus.checkAll() {
			log.Printf("                                                                                                          ！！当前运行模式为：[ 客户端 ] 模式！！")
			go Pholcus.Client()
		}
	// case status.OFFLINE:
	// 	fallthrough
	default:
		log.Printf("                                                                                                          ！！当前运行模式为：[ 单机 ] 模式！！")
		return
	}

	// 循环处理请求
	go Pholcus.reqHandle()
}

// 生成task并添加至库，服务器模式专用
func (self *Node) AddNewTask(spiders []string, keywords string) {
	t := &task.Task{}

	t.Spiders = spiders
	t.Keywords = keywords

	// 从配置读取字段
	t.ThreadNum = cache.Task.ThreadNum
	t.BaseSleeptime = cache.Task.BaseSleeptime
	t.RandomSleepPeriod = cache.Task.RandomSleepPeriod
	t.OutType = cache.Task.OutType
	t.DockerCap = cache.Task.DockerCap
	t.DockerQueueCap = cache.Task.DockerQueueCap
	t.MaxPage = cache.Task.MaxPage

	// 存入
	self.tasks.Push(t)
	// log.Printf(" *     [新增任务]   详情： %#v", *t)
}

// 请求任务，客户端模式专用，网络重连后重新发送请求
func (self *Node) DownTask() *task.Task {
startLabel:
	self.WaitConn()

	if len(self.tasks.Ready) == 0 {
		go cache.PushNetData(status.REQTASK, nil, "")
	}

	for len(self.tasks.Ready) == 0 {
		if !self.HasConn() {
			goto startLabel
		}
		time.Sleep(5e7)
		continue
	}

	return self.tasks.Pull()
}

func (self *Node) reqHandle() {
	for {
		data := <-cache.ReceiveDocker
		switch data.Type {
		case status.REQTASK:
			self.sendTask(data)
			self.GetConn(data.From).Unblock()
			// log.Println("请求处理完成，已解除服务器读取循环的阻塞")
		case status.TASK:
			self.receiveTask(data)
		}
	}
}

// 分发任务
func (self *Node) sendTask(data *cache.NetData) {
	var t task.Task
	var ok bool
	for {
		if t, ok = self.tasks.Out(data.From, len(self.Conns)); ok {
			break
		}
		time.Sleep(1e9)
	}
	cache.PushNetData(status.TASK, t, data.From)
	self.AutoSend(<-cache.SendDocker)
}

// 将接收来的任务加入库
func (self *Node) receiveTask(data *cache.NetData) {
	// log.Println("将任务入库", data)
	d, err := json.Marshal(data.Body)
	if err != nil {
		log.Println("json编码失败", data.Body)
		return
	}
	t := &task.Task{}
	err = json.Unmarshal(d, t)
	if err != nil {
		log.Println("json解码失败", data.Body)
		return
	}
	self.tasks.Into(t)
}

func (self *Node) checkPort() bool {
	if cache.Task.Port == 0 {
		log.Println(" *     —— 亲，分布式端口不能为空哦~")
		return false
	}
	return true
}

func (self *Node) checkAll() bool {
	if cache.Task.Master == "" || !self.checkPort() {
		log.Println(" *     —— 亲，服务器地址不能为空哦~")
		return false
	}
	return true
}
