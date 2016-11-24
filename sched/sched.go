package sched

import (
	"container/heap"
	"time"
	"fmt"
)

type op struct{
	do func()
	scheduleTime time.Time
	interval time.Duration
}

type opHeap []*op
func (oh opHeap) Len() int { return len(oh) }
func (oh opHeap) Swap(i, j int) { oh[i], oh[j] = oh[j], oh[i] }

func (oh opHeap) Less(i, j int) bool {
	return oh[i].scheduleTime.Before(oh[j].scheduleTime)
}

func (oh *opHeap) Push(x interface{}) {
	*oh = append(*oh, x.(*op))
}

func (oh *opHeap) Pop() interface{} {
	old := *oh
	n := len(old)
	x := old[n-1]
	*oh = old[0 : n-1]
	return x
}


type Scheduler struct{
	heap *opHeap
	stop bool
	anythingScheduled bool
}

func New() *Scheduler {
	oh := &opHeap{}
	heap.Init(oh)

	return &Scheduler{
		heap: oh,
	}
}

func (s *Scheduler) Schedule(do func(), initialTime time.Time, interval time.Duration) {
	o := &op{
		do: do,
		scheduleTime: initialTime,
		interval: interval,
	}
	s.anythingScheduled = true
	heap.Push(s.heap, o)
}

func (s *Scheduler) Run() error {

	if !s.anythingScheduled {
		return fmt.Errorf("Nothing scheduled")
	}

	s.stop = false

	for {

		o := heap.Pop(s.heap).(*op)
		time.Sleep(o.scheduleTime.Sub(time.Now()))
		if s.stop {
			break
		}
		o.scheduleTime = o.scheduleTime.Add(o.interval)
		heap.Push(s.heap, o)
		o.do()

	}

	return nil
}

func (s *Scheduler) Stop() {
	s.stop = true
}

