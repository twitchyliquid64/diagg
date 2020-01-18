package flowui

import (
	"sync"
	"time"
)

type averageMetric struct {
	Last     [25]int64
	Name     string
	Category string

	cursor int
	lock   sync.Mutex
}

func (m *averageMetric) Time(started time.Time) {
	m.lock.Lock()
	defer m.lock.Unlock()
	now := time.Now()
	m.Last[m.cursor] = int64(now.Sub(started))
	m.cursor = (m.cursor + 1) % 25
}

func (m *averageMetric) Metric() string {
	return m.Name
}
func (m *averageMetric) Compute() string {
	m.lock.Lock()
	defer m.lock.Unlock()
	var sum int64
	for _, v := range m.Last {
		sum += v
	}
	avg := sum / 25
	return time.Duration(avg).String()
}
