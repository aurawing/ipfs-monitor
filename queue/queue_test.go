package queue

import "testing"

var syncQueue = NewSyncQueue()

func TestSaveAndRecover(t *testing.T) {
	syncQueue.Push("test")
	err := syncQueue.Save("test")
	if err != nil {
		t.Error(err)
	}
	item := syncQueue.Pop()
	t.Log(item.(string))
	syncQueue.Recover("test")
	item = syncQueue.Pop()
	if item.(string) == "test" {
		t.Log("测试通过")
	} else {
		t.Error("测试未通过")
	}
}

func TestLength(t *testing.T) {
	if syncQueue.Length() == 0 {
		t.Log("队列为空")
	}
	syncQueue.Push("test")
	if syncQueue.Length() == 1 {
		t.Logf("队列长度为%d", syncQueue.Length())
	}
}
