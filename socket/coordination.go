package socket

var _ Parent = &coordination{}
var _ Child = &coordination{}

type coordination struct {
	n    int
	stop chan struct{}
	done chan struct{}
}

type Parent interface {
	Wait()
}

type Child interface {
	ShouldStop() chan struct{}
	Done()
}

func NewCoordination(n int) (Parent, Child) {
	c := &coordination{
		n,
		make(chan struct{}),
		make(chan struct{}),
	}
	return Parent(c), Child(c)
}

func (c *coordination) Done() {
	c.done <- struct{}{}
}

func (c *coordination) ShouldStop() chan struct{} {
	return c.stop
}

func (c *coordination) Wait() {
	for i := 0; i < c.n; i++ {
		c.stop <- struct{}{}
	}

	for i := 0; i < c.n; i++ {
		<-c.done
	}
}
