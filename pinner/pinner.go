package pinner

import (
	"ipfs-monitor/command"
	"ipfs-monitor/queue"
	"log"
	"os"
)

var syncQueue = queue.NewSyncQueue()

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func PinAsync(hashs []string) {
	for _, hash := range hashs {
		syncQueue.Push(hash)
	}
}

func PinService() {
	go func() {
		for {
			hash := syncQueue.Pop()
			stdlog.Println("Pinning file: ", hash)
			_, err := command.PinFile(hash.(string))
			if err != nil {
				errlog.Printf("Pin file %s failed, error: %s\n", hash, err)
				continue
			}
			stdlog.Printf("Pin file %s successed.\n", hash)
		}
	}()
}
