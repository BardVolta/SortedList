package clist

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type SortedList interface {
	// Contains 检查一个元素是否存在，如果存在则返回 true，否则返回 false
	Contains(value int) bool

	// Insert 插入一个元素，如果此操作成功插入一个元素，则返回 true，否则返回 false
	Insert(value int) bool

	// Delete 删除一个元素，如果此操作成功删除一个元素，则返回 true，否则返回 false
	Delete(value int) bool

	// Range 遍历此有序链表的所有元素，如果 f 返回 false，则停止遍历
	Range(f func(value int) bool)

	// Len 返回有序链表的元素个数
	Len() int
}

type intNode struct {
	value  int
	marked int32
	next   *intNode
	mu     sync.Mutex
}

type IntList struct {
	head   *intNode
	length int64
}

func newIntNode(value int) *intNode {
	return &intNode{value: value}
}

func NewInt() *IntList {
	return &IntList{head: newIntNode(0)}
}

func (l *IntList) Insert(value int) bool {
	var a, b *intNode
	a = l.head
	b = a.getNextNode()
	for b != nil && b.value < value {
		a = b
		b = b.getNextNode()
	}

	// Check if the node is exist.
	if b != nil && b.value == value {
		return false
	}

	// Critical section
	a.mu.Lock()

	if a.next != b || atomic.LoadInt32(&a.marked) == 1 {
		a.mu.Unlock()
		return l.Insert(value)
	}

	x := newIntNode(value)
	x.storeNextNode(b)
	a.storeNextNode(x)
	atomic.AddInt64(&l.length, 1)

	a.mu.Unlock()

	return true
}

func (l *IntList) Delete(value int) bool {
	var a, b *intNode
	a = l.head
	b = a.getNextNode()

	for b != nil && b.value < value {
		a = b
		b = b.getNextNode()
	}

	// Check if b is not exists
	if b == nil || b.value != value {
		return false
	}

	// Critical section
	b.mu.Lock()
	if atomic.LoadInt32(&b.marked) == 1 {
		b.mu.Unlock()
		return l.Delete(value)
	}
	a.mu.Lock()

	if a.next != b || atomic.LoadInt32(&a.marked) == 1 {
		a.mu.Unlock()
		b.mu.Unlock()
		return l.Delete(value)
	}

	a.storeNextNode(b.getNextNode())
	atomic.StoreInt32(&b.marked, 1)

	atomic.AddInt64(&l.length, -1)

	a.mu.Unlock()
	b.mu.Unlock()
	return true
}

// Contains return the value in the list
func (l *IntList) Contains(value int) bool {
	x := l.head.getNextNode()
	for x != nil && x.value < value {
		x = x.getNextNode()
	}
	if x == nil {
		return false
	}
	return x.value == value && atomic.LoadInt32(&x.marked) == 0
}

// Range iteration the function
func (l *IntList) Range(f func(value int) bool) {
	x := l.head.getNextNode()
	for x != nil {
		if !f(x.value) {
			break
		}
		x = x.getNextNode()
	}
}

// Len returns the length of the list
func (l *IntList) Len() int {
	return int(atomic.LoadInt64(&l.length))
}

// getNextNode same with `n.next`(atomic)
func (n *intNode) getNextNode() *intNode {
	return (*intNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next))))
}

// storeNextNode same with `n.next = node`(atomic)
func (n *intNode) storeNextNode(node *intNode) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&n.next)), unsafe.Pointer(node))
}
