package pinner

import (
	"os"
	"testing"
	"time"
)

func TestSaveAndReadPinningList(t *testing.T) {
	// pinningList := [1]*Task{}
	task := "testhash"
	SavePinningTask(0, task)
	t.Log(pinningList)
	pinningList[0] = "test"
	ReadPinningTask()
	if pinningList[0] == "testhash" {
		t.Log("测试过")
	} else {
		t.Error(pinningList)
	}
	for _, v := range pinningList {
		t.Log(v)
	}
	os.Remove(".pinningList")
}

func TestPinSync(t *testing.T) {
	pinningList = [2]string{"1", "3"}
	PinAsync(pinningList[:])
	if pinningCount == 2 {
		t.Log("任务数量正确")
	}
	if syncQueue.Length() == 2 {
		t.Log("队列长度正确")
	}
	for i := 0; i < 2; i++ {
		item := syncQueue.Pop().(Task)
		if pinningList[i] == item.Hash {
			t.Log("第" + string(i) + "个元素正确")
		} else {
			t.Log("第" + string(i) + "个元素错误")
		}
	}
}

func TestHandleResult(t *testing.T) {
	resQueue.Push(&Task{"1", 3, 1})
	resQueue.Push(&Task{"2", 0, 2})
	resQueue.Push(&Task{"QmWrKAwF86yEWSHTQmeF6aHDfpYtuwAwCizyJi4RfDW9Yb", 2, 1})
	resQueue.Push(&Task{"0", 3, 0})
	go handleTask(0)
	go handleTask(1)
	go handleResults()
	time.Sleep(30 * time.Second)
	if len(FailList) == 2 && syncQueue.Length() == 0 {
		t.Log("错误列表长度正确")
	} else {
		item := syncQueue.Pop()
		task := item.(Task)
		t.Error("错误列表长度错误", FailList, syncQueue.Length(), task.Hash)
	}
}

func TestPinSevice(t *testing.T) {
	PinAsync([]string{"1", "2", "QmWrKAwF86yEWSHTQmeF6aHDfpYtuwAwCizyJi4RfDW9Yb"})
	PinService()
	time.Sleep(30 * time.Second)
	if len(FailList) != 2 {
		t.Error(FailList)
	}

}
