package pinner

import (
	"io/ioutil"
	"ipfs-monitor/command"
	"ipfs-monitor/config"
	"ipfs-monitor/queue"
	"log"
	"os"
	"sync"
)

var JobCount int

var lock sync.Mutex

var pinningCount int

var syncQueue = queue.NewSyncQueue() // 待处理任务队列
var resQueue = queue.NewSyncQueue()  //处理结果队列

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

type Task struct {
	Hash         string
	TimeoutCount int //超时计数
	Status       int //0 正常，1 下载超时 2，失败
}

var FailList []FailItem

// Faliure history
type FailItem struct {
	Hash   string
	Code   int
	Detail string
}

func PinAsync(hashs []string) {
	for _, hash := range hashs {
		// 生成一个状态正常的任务
		task := Task{hash, 0, 0}
		if !syncQueue.Has(task) {
			lock.Lock()
			syncQueue.Push(task)
			pinningCount++
			lock.Unlock()
		}
	}
}

func PinningFileSize() int {
	return pinningCount
}

func GetNeedTaskNum() int {
	return config.GetMaxTaskNum() - pinningCount
}

func PinService() {
	for i := 0; i < JobCount; i++ {
		go func() {
			for {
				item := syncQueue.Pop()
				task := item.(Task)
				var progress int64
				err := command.GetFile(task.Hash, ioutil.Discard, func(reads int64, total int64) {
					if (100*reads/total - progress) >= 5 {
						progress = 100 * reads / total
						stdlog.Printf("File: %s has downloaded %d", task.Hash, progress)
						peerId, _ := command.GetPeerID()
						command.IpfsPub(peerId+"-download-"+task.Hash+"log", string(progress))
					}
				})
				if err != nil {
					switch err.Error() {
					case "Download time out":
						task.TimeoutCount++
						task.Status = 1
					case "Request time out":
						task.Status = 2
					default:
						task.TimeoutCount++
						task.Status = 1
					}
				} else {
					stdlog.Println("Pinning file: ", task, GetNeedTaskNum(), PinningFileSize())
					_, err = command.PinFile(task.Hash)
					if err != nil {
						errlog.Printf("Pin file %s failed, error: %s\n", task.Hash, err)
						task.Status = 2
					} else {
						stdlog.Printf("Pin file %s successed.\n", task.Hash)
						task.Status = 0
					}
				}
				resQueue.Push(&task)
			}
		}()

	}
	// 处理pin结果
	go func() {
		for {
			item := resQueue.Pop()
			task := item.(*Task)
			if task.Status == 0 || task.Status == 2 || task.TimeoutCount >= 3 {
				lock.Lock()
				pinningCount--
				lock.Unlock()
				if task.Status == 2 {
					FailList = append(FailList, FailItem{task.Hash, 1, "Request time out"})
					errlog.Printf("任务：%s 连接矿工失败 返回失败", task.Hash)
				} else if task.Status == 1 {
					FailList = append(FailList, FailItem{task.Hash, 1, "Download time out"})
					errlog.Printf("任务：%s 尝试下载失败3次 返回失败", task.Hash)
				}
			} else {
				errlog.Printf("任务：%s 尝试下载失败%d次", task.Hash, task.TimeoutCount)
				syncQueue.Push(*task)
			}
		}
	}()
}
