package queue

import (
	"io/ioutil"
	"sync"
	"unsafe"

	"gopkg.in/eapache/queue.v1"
)

// Synchronous FIFO queue
type SyncQueue struct {
	lock    sync.Mutex
	popable *sync.Cond
	buffer  *queue.Queue
	closed  bool
}

// Create a new SyncQueue
func NewSyncQueue() *SyncQueue {
	ch := &SyncQueue{
		buffer: queue.New(),
	}
	ch.popable = sync.NewCond(&ch.lock)
	return ch
}

// Pop an item from SyncQueue, will block if SyncQueue is empty
func (q *SyncQueue) Pop() (v interface{}) {
	c := q.popable
	buffer := q.buffer

	q.lock.Lock()
	for buffer.Length() == 0 && !q.closed {
		c.Wait()
	}

	if buffer.Length() > 0 {
		v = buffer.Peek()
		buffer.Remove()
	}

	q.lock.Unlock()
	return
}

// Try to pop an item from SyncQueue, will return immediately with bool=false if SyncQueue is empty
func (q *SyncQueue) TryPop() (v interface{}, ok bool) {
	buffer := q.buffer

	q.lock.Lock()

	if buffer.Length() > 0 {
		v = buffer.Peek()
		buffer.Remove()
		ok = true
	} else if q.closed {
		ok = true
	}

	q.lock.Unlock()
	return
}

// Push an item to SyncQueue. Always returns immediately without blocking
func (q *SyncQueue) Push(v interface{}) {
	q.lock.Lock()
	if !q.closed {
		q.buffer.Add(v)
		q.popable.Signal()
	}
	q.lock.Unlock()
}

// Get the length of SyncQueue
func (q *SyncQueue) Len() (l int) {
	q.lock.Lock()
	l = q.buffer.Length()
	q.lock.Unlock()
	return
}

func (q *SyncQueue) Has(v interface{}) bool {
	q.lock.Lock()
	for i := 0; i < q.buffer.Length(); i++ {
		if v == q.buffer.Get(i) {
			return true
		}
	}
	q.lock.Unlock()
	return false
}

func (q *SyncQueue) Close() {
	q.lock.Lock()
	if !q.closed {
		q.closed = true
		q.popable.Signal()
	}
	q.lock.Unlock()
}

type SliceMock struct {
	addr uintptr
	len  int
	cap  int
}

func Encode(data *queue.Queue) []byte {
	len := unsafe.Sizeof(*data)
	bts := &SliceMock{
		addr: uintptr(unsafe.Pointer(data)),
		cap:  int(len),
		len:  int(len),
	}
	return *(*[]byte)(unsafe.Pointer(bts))
}

func Decode(data []byte) *queue.Queue {
	return *(**queue.Queue)(unsafe.Pointer(&data))
}

func (q *SyncQueue) Save(filename string) error {
	buf := Encode(q.buffer)
	return ioutil.WriteFile(filename, buf, 0644)
}

func (q *SyncQueue) Recover(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	q.buffer = Decode(buf)
	return nil
}

func (q *SyncQueue) Length() int {
	return q.buffer.Length()
}
