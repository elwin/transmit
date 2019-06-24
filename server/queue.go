package server

import "github.com/elwin/transmit/striping"

type StripingQueue interface {
	Enqueue(segment *striping.Segment)
	Dequeue() *striping.Segment
	Len() int
}

var _ StripingQueue = StripingQueueImplementation{}

// This implementation is NOT thread safe
type StripingQueueImplementation struct {
	first *Element
	len   int
}

type Element struct {
	value      *striping.Segment
	prev, next *Element
}

func (queue StripingQueueImplementation) Enqueue(segment *striping.Segment) {
	e := &Element{segment, nil, nil}

	if queue.first == nil {
		queue.first = e
	} else {
		cur := queue.first
		for cur.next != nil {
			cur = cur.next
		}

		cur.next = e
		e.prev = cur
	}

	queue.len++
}

func (queue StripingQueueImplementation) Dequeue() *striping.Segment {
	e := queue.first
	queue.first = e.next
	queue.first.prev = nil
	e.next = nil
	return e.value
}

func (queue StripingQueueImplementation) Len() int {
	return queue.len
}
