package pinner

import (
	"ipfs-monitor/command"
	"ipfs-monitor/queue"
	"log"
	"os"
)

var syncQueue = queue.NewSyncQueue()

var pinning bool = false

var stdlog, errlog *log.Logger

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func PinAsync(hashs []string) {
	for _, hash := range hashs {
		if !syncQueue.Has(hash) {
			syncQueue.Push(hash)
		}
	}
}

func PinningFileSize() uint32 {
	if syncQueue.Len() > 0 {
		return uint32(syncQueue.Len() + 1)
	} else {
		if pinning {
			return 1
		} else {
			return 0
		}
	}
}

func PinService() {
	go func() {
		for {
			hash := syncQueue.Pop()
			pinning = true
			stdlog.Println("Pinning file: ", hash)
			_, err := command.PinFile(hash.(string))
			pinning = false
			if err != nil {
				errlog.Printf("Pin file %s failed, error: %s\n", hash, err)
				continue
			}
			stdlog.Printf("Pin file %s successed.\n", hash)
		}
	}()
}
