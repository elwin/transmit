package striping

import "github.com/elwin/transmit/queue"

type SegmentQueue struct {
	Internal queue.Queue
}

var _ queue.Sortable = &Item{}

type Item struct {
	Segment
}

func (item *Item) Less(b queue.Sortable) bool {
	return item.OffsetCount < b.(*Item).OffsetCount
}

func (q *SegmentQueue) Push(segment *Segment) {
	q.Internal.Push(segment)
}

func (q *SegmentQueue) Pop() *Segment {
	return q.Internal.Pop().(*Segment)
}

func (q *SegmentQueue) Peek() *Segment {
	return q.Internal.Peek().(*Segment)
}

func (q *SegmentQueue) Len() int {
	return q.Internal.Len()
}

func NewSegmentQueue() *SegmentQueue {
	return &SegmentQueue{queue.NewQueue()}
}
