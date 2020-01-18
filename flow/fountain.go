package flow

import (
	"fmt"
	"sync"
)

var (
	nodeNum int
	padNum  int
	edgeNum int
	l       sync.Mutex
)

func AllocNodeID(t string) string {
	l.Lock()
	defer l.Unlock()

	nodeNum++
	if t == "" {
		return fmt.Sprintf("node-%d", nodeNum-1)
	}
	return fmt.Sprintf("node-%s-%d", t, nodeNum-1)
}

func AllocPadID(t string) string {
	l.Lock()
	defer l.Unlock()

	padNum++
	if t == "" {
		return fmt.Sprintf("pad-%d", padNum-1)
	}
	return fmt.Sprintf("pad-%s-%d", t, padNum-1)
}

func AllocEdgeID(t string) string {
	l.Lock()
	defer l.Unlock()

	edgeNum++
	if t == "" {
		return fmt.Sprintf("edge-%d", edgeNum-1)
	}
	return fmt.Sprintf("edge-%s-%d", t, edgeNum-1)
}
