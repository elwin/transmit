package server

import "github.com/elwin/transmit/striping"

func partitionData(data []byte, maxSize int) []*striping.Segment {

	/*
		segments := []*striping.Segment{
			striping.NewSegment(data[:2], 0),
			striping.NewSegment(data[2:5], 2),
			striping.NewSegment(data[5:6], 5),
			striping.NewSegment(data[6:10], 6),
			striping.NewSegment(data[10:], 10),
		}
	*/

	var segments []*striping.Segment

	for i := 0; i < len(data); i += maxSize {

		start := i
		end := i + maxSize
		if end > len(data) {
			end = len(data)
		}

		segment := striping.NewSegment(data[start:end], start)
		segments = append(segments, segment)

	}

	return segments
}

func distributeSegments(segments []*striping.Segment, n int) []SegmentQueue {

	var queue []SegmentQueue
	for i := 0; i < n; i++ {
		queue = append(queue, NewSegmentQueue())
	}

	for i, segment := range segments {
		queue[i%n].Enqueue(segment)
	}

	return queue
}
