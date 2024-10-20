package network

import (
	"sync"
	"testing"
)

func TestStartNewNode(t *testing.T) {

	var wg sync.WaitGroup

	for i := 1; i <= 4; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			nodeId := i
			node := StartNewNode(nodeId, false)

			if node.NodeID != nodeId {
				t.Errorf("expected NodeID to be %d, got %d", nodeId, node.NodeID)
			}
		}(i)
	}
	wg.Wait()
}
