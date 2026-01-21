package buffer

import "sync"

// RingBuffer is a thread-safe generic circular buffer.
type RingBuffer[T any] struct {
	mu       sync.RWMutex
	data     []T
	head     int // next write position
	count    int // number of elements
	capacity int
}

// NewRingBuffer creates a new ring buffer with the given capacity.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		data:     make([]T, capacity),
		capacity: capacity,
	}
}

// Push adds an item to the buffer. If the buffer is full, the oldest item is overwritten.
func (rb *RingBuffer[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.head] = item
	rb.head = (rb.head + 1) % rb.capacity

	if rb.count < rb.capacity {
		rb.count++
	}
}

// Len returns the number of elements in the buffer.
func (rb *RingBuffer[T]) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// Capacity returns the maximum capacity of the buffer.
func (rb *RingBuffer[T]) Capacity() int {
	return rb.capacity
}

// Get returns the item at the given index (0 is oldest).
func (rb *RingBuffer[T]) Get(index int) (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var zero T
	if index < 0 || index >= rb.count {
		return zero, false
	}

	start := (rb.head - rb.count + rb.capacity) % rb.capacity
	actualIndex := (start + index) % rb.capacity
	return rb.data[actualIndex], true
}

// GetLast returns the most recent item.
func (rb *RingBuffer[T]) GetLast() (T, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var zero T
	if rb.count == 0 {
		return zero, false
	}

	lastIndex := (rb.head - 1 + rb.capacity) % rb.capacity
	return rb.data[lastIndex], true
}

// GetRange returns items from start to end index (inclusive, 0 is oldest).
func (rb *RingBuffer[T]) GetRange(start, end int) []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if end >= rb.count {
		end = rb.count - 1
	}
	if start > end || rb.count == 0 {
		return nil
	}

	result := make([]T, end-start+1)
	bufStart := (rb.head - rb.count + rb.capacity) % rb.capacity

	for i := start; i <= end; i++ {
		actualIndex := (bufStart + i) % rb.capacity
		result[i-start] = rb.data[actualIndex]
	}

	return result
}

// GetLastN returns the last n items (most recent last).
func (rb *RingBuffer[T]) GetLastN(n int) []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n > rb.count {
		n = rb.count
	}
	if n <= 0 {
		return nil
	}

	result := make([]T, n)
	bufStart := (rb.head - n + rb.capacity) % rb.capacity

	for i := 0; i < n; i++ {
		actualIndex := (bufStart + i) % rb.capacity
		result[i] = rb.data[actualIndex]
	}

	return result
}

// All returns all items in the buffer (oldest first).
func (rb *RingBuffer[T]) All() []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]T, rb.count)
	start := (rb.head - rb.count + rb.capacity) % rb.capacity

	for i := 0; i < rb.count; i++ {
		actualIndex := (start + i) % rb.capacity
		result[i] = rb.data[actualIndex]
	}

	return result
}

// Clear removes all items from the buffer.
func (rb *RingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.count = 0
}
