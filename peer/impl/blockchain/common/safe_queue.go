package common

import (
	"container/list"
	"sync"
)

// SafeQueue is a thread-safe version of queue
type SafeQueue[T any] struct {
	q  *list.List
	mu sync.Mutex
}

// NewSafeQueue is not used so far
func NewSafeQueue[T any]() SafeQueue[T] {
	return SafeQueue[T]{
		q:  list.New(),
		mu: sync.Mutex{},
	}
}

func (c *SafeQueue[T]) Len() uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	return uint(c.q.Len())
}

func (c *SafeQueue[T]) Enqueue(v T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.q.PushBack(v)
}

func (c *SafeQueue[T]) Front() T {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.q.Front().Value.(T)
}

func (c *SafeQueue[T]) Dequeue() T {
	c.mu.Lock()
	defer c.mu.Unlock()
	i := c.q.Front().Value
	c.q.Remove(c.q.Front())
	return i.(T)
}

func (c *SafeQueue[T]) IsEmpty() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.q.Len() == 0
}
