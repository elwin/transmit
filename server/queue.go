package server

import "github.com/elwin/transmit/striping"

type SegmentQueue interface {
	Enqueue(segment *striping.Segment)
	Dequeue() *striping.Segment
	Len() int
	Empty() bool
}

var _ SegmentQueue = &SegmentQueueImplementation{}

// This implementation is NOT thread safe
type SegmentQueueImplementation struct {
	first *Element
	len   int
}

func (queue *SegmentQueueImplementation) Empty() bool {
	return !(queue.len > 0)
}

type Element struct {
	value      *striping.Segment
	prev, next *Element
}

func (queue *SegmentQueueImplementation) Enqueue(segment *striping.Segment) {
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

func (queue *SegmentQueueImplementation) Dequeue() *striping.Segment {

	// Maybe return error
	if queue.first == nil {
		panic("Cannot dequeue from empty queue")
	}

	e := queue.first
	queue.first = e.next
	if queue.first != nil {
		queue.first.prev = nil
	}
	e.next = nil

	queue.len--

	return e.value
}

func (queue *SegmentQueueImplementation) Len() int {
	return queue.len
}

func NewSegmentQueue() SegmentQueue {
	return &SegmentQueueImplementation{}
}
