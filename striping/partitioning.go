package striping

func PartitionData(data []byte, maxSize int) []*Segment {

	var segments []*Segment

	for i := 0; i < len(data); i += maxSize {

		start := i
		end := i + maxSize
		if end > len(data) {
			end = len(data)
		}

		segment := NewSegment(data[start:end], start)
		segments = append(segments, segment)

	}

	return segments
}

func DistributeSegments(segments []*Segment, n int) []SegmentQueue {

	var queue []SegmentQueue
	for i := 0; i < n; i++ {
		queue = append(queue, NewSegmentQueue())
	}

	for i, segment := range segments {
		queue[i%n].Enqueue(segment)
	}

	return queue
}
