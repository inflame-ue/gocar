package scheduler

// ReadyQueue implements the [[heap.Interface]].
// Each item in the ReadyQueue is assigned a priority level.
type ReadyQueue struct {
	level map[string]int
	items []string
}

func NewReadyQueue(level map[string]int) *ReadyQueue {
	return &ReadyQueue{level: level}
}

func (q *ReadyQueue) Len() int {
	return len(q.items)
}

func (q *ReadyQueue) Less(i, j int) bool {
	a, b := q.items[i], q.items[j]
	if q.level[a] != q.level[b] {
		return q.level[a] > q.level[b]
	}
	return a < b // deterministic tiebreak
}

func (q *ReadyQueue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *ReadyQueue) Push(x any) {
	q.items = append(q.items, x.(string))
}

func (q *ReadyQueue) Pop() any {
	old := q.items
	n := len(old)
	x := old[n-1]
	q.items = old[:n-1]
	return x
}
