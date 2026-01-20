package buffer

import (
	"sync"
	"testing"
)

func TestRingBuffer_Basic(t *testing.T) {
	rb := NewRingBuffer[int](5)

	if rb.Len() != 0 {
		t.Errorf("expected len 0, got %d", rb.Len())
	}

	// Push some items
	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	if rb.Len() != 3 {
		t.Errorf("expected len 3, got %d", rb.Len())
	}

	// Get items
	v, ok := rb.Get(0)
	if !ok || v != 1 {
		t.Errorf("expected 1, got %d, ok=%v", v, ok)
	}

	v, ok = rb.Get(2)
	if !ok || v != 3 {
		t.Errorf("expected 3, got %d, ok=%v", v, ok)
	}

	// Get out of bounds
	_, ok = rb.Get(5)
	if ok {
		t.Error("expected not ok for out of bounds")
	}
}

func TestRingBuffer_Wraparound(t *testing.T) {
	rb := NewRingBuffer[int](3)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	rb.Push(4) // overwrites 1
	rb.Push(5) // overwrites 2

	if rb.Len() != 3 {
		t.Errorf("expected len 3, got %d", rb.Len())
	}

	// Should contain 3, 4, 5
	all := rb.All()
	expected := []int{3, 4, 5}
	for i, v := range expected {
		if all[i] != v {
			t.Errorf("expected %d at index %d, got %d", v, i, all[i])
		}
	}
}

func TestRingBuffer_GetLast(t *testing.T) {
	rb := NewRingBuffer[int](5)

	_, ok := rb.GetLast()
	if ok {
		t.Error("expected not ok for empty buffer")
	}

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	v, ok := rb.GetLast()
	if !ok || v != 3 {
		t.Errorf("expected 3, got %d, ok=%v", v, ok)
	}
}

func TestRingBuffer_GetLastN(t *testing.T) {
	rb := NewRingBuffer[int](5)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	rb.Push(4)
	rb.Push(5)

	last3 := rb.GetLastN(3)
	expected := []int{3, 4, 5}
	for i, v := range expected {
		if last3[i] != v {
			t.Errorf("expected %d at index %d, got %d", v, i, last3[i])
		}
	}

	// Request more than available
	all := rb.GetLastN(10)
	if len(all) != 5 {
		t.Errorf("expected 5 items, got %d", len(all))
	}
}

func TestRingBuffer_GetRange(t *testing.T) {
	rb := NewRingBuffer[int](5)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	rb.Push(4)
	rb.Push(5)

	// Get middle range
	r := rb.GetRange(1, 3)
	expected := []int{2, 3, 4}
	for i, v := range expected {
		if r[i] != v {
			t.Errorf("expected %d at index %d, got %d", v, i, r[i])
		}
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb := NewRingBuffer[int](5)

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	rb.Clear()

	if rb.Len() != 0 {
		t.Errorf("expected len 0 after clear, got %d", rb.Len())
	}

	_, ok := rb.GetLast()
	if ok {
		t.Error("expected not ok after clear")
	}
}

func TestRingBuffer_Concurrent(t *testing.T) {
	rb := NewRingBuffer[int](100)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				rb.Push(base*100 + j)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = rb.Len()
				_ = rb.All()
				_, _ = rb.GetLast()
			}
		}()
	}

	wg.Wait()

	// Should not panic and len should be at most capacity
	if rb.Len() > rb.Capacity() {
		t.Errorf("len %d exceeds capacity %d", rb.Len(), rb.Capacity())
	}
}
