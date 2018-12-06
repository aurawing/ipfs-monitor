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

var syncQueue = queue.NewSyncQueue()

// var pinningQueue = queue.NewSyncQueue()

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
	// _, err := os.Stat(".pinningFile")
	// if err == nil {
	// 	syncQueue.Recover(".pinningFile")
	// }
}

type Task struct {
	Hash         string
	TimeoutCount int //超时计数
	Status       int //0 get中，1 成功 2，失败
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
		task := Task{hash, 0, 0}
		if !syncQueue.Has(task) {
			lock.Lock()
			syncQueue.Push(task)
			pinningCount++
			lock.Unlock()
		}
	}
}

func RetryPinAsync(task *Task) *Task {
	if !syncQueue.Has(*task) {
		task.TimeoutCount += 1
		if task.TimeoutCount < 3 {
			errlog.Printf("任务 %s 超时 尝试第%d次\n", task.Hash, task.TimeoutCount+1)
			return task
		} else {
			FailList = append(FailList, FailItem{task.Hash, 2, "Download time out"})
			errlog.Printf("任务 %s 失败 返回失效\n", task.Hash)
		}
	}
	return nil
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
				// pinningQueue.Push(&task)
				// pinningQueue.Save(".pinningList")
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
						RetryPinAsync(&task)
					case "Request time out":
						FailList = append(FailList, FailItem{task.Hash, 1, "Request time out"})
						errlog.Printf("任务 %s 失败 矿工节点连接超时, error: %s\n", task.Hash, err)
					default:
						RetryPinAsync(&task)
					}
				} else {
					stdlog.Println("Pinning file: ", task, GetNeedTaskNum(), PinningFileSize())
					_, err = command.PinFile(task.Hash)
					if err != nil {
						errlog.Printf("Pin file %s failed, error: %s\n", task.Hash, err)
					} else {
						stdlog.Printf("Pin file %s successed.\n", task.Hash)
					}
				}
				lock.Lock()
				pinningCount--
				lock.Unlock()
			}
		}()
	}
}
