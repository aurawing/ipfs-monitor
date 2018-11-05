package pinner

import (
	"io/ioutil"
	"ipfs-monitor/command"
	"ipfs-monitor/queue"
	"log"
	"os"
	"sync"
)

var JobCount int

var lock sync.Mutex

var pinningCount uint32

var syncQueue = queue.NewSyncQueue()

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func PinAsync(hashs []string) {
	for _, hash := range hashs {
		if !syncQueue.Has(hash) {
			lock.Lock()
			syncQueue.Push(hash)
			pinningCount++
			lock.Unlock()
		}
	}
}

func PinningFileSize() uint32 {
	return pinningCount
}

func PinService() {
	for i := 0; i < JobCount; i++ {
		go func() {
			for {
				hash := syncQueue.Pop()
				var progress int64
				err := command.GetFile(hash.(string), ioutil.Discard, func(reads int64, total int64) {
					if (100*reads/total - progress) >= 5 {
						progress = 100 * reads / total
						stdlog.Printf("File: %s has downloaded %d", hash, progress)
					}

				})
				if err != nil {
					errlog.Printf("Get file %s failed, error: %s\n", hash, err)
				} else {
					stdlog.Println("Pinning file: ", hash)
					_, err = command.PinFile(hash.(string))
					if err != nil {
						errlog.Printf("Pin file %s failed, error: %s\n", hash, err)
					} else {
						stdlog.Printf("Pin file %s successed.\n", hash)
					}
				}
				lock.Lock()
				pinningCount--
				lock.Unlock()
			}
		}()
	}
}
