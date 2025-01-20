package node

import (
	"deukyunlee/hotstuff/block"
	"deukyunlee/hotstuff/message"
	"time"
)

func (n *Node) HandleDecide(msg message.Message) {
	logger.Infof("Received Decide for block %s. Finalizing...\n",msg.Block.Hash)

	logger.Infof("Block %d, hash: %s is now decided. Current view: %d\n",msg.Block.Number, msg.Block.Hash, n.View)

	if msg.View >= n.View {
		n.View = msg.View + 1
	}

	newBlock := block.CreateBlock(msg.Block, "")
	newMsg := message.Message{
		Type:     message.Prepare,
		Block:    newBlock,
		SenderID: n.ID,
		View:     n.View,
	}

	n.Committed = append(n.Committed, msg.Block)

	if n.IsLeaderNode() {
		n.Propose(newMsg)
	}

	currentTime := time.Now()
	interval := currentTime.Sub(n.LastBlockTime)
	
	n.RecentIntervals = append(n.RecentIntervals, interval)
	if len(n.RecentIntervals) > n.WindowSize {
		n.RecentIntervals = n.RecentIntervals[1:]
	}
	
	n.TotalInterval += interval
	totalAvg := n.TotalInterval.Seconds() / float64(n.BlockCount + 1)
	
	var recentAvg float64
	if len(n.RecentIntervals) > 0 {
		var sum time.Duration
		for _, i := range n.RecentIntervals {
			sum += i
		}
		recentAvg = sum.Seconds() / float64(len(n.RecentIntervals))
	}
	
	logger.Infof("- Current interval: %.5fs", interval.Seconds())
	logger.Infof("[blockNo: %d] Current Interval: %.5fs, Total average: %.5fs",n.interval.Seconds(), totalAvg)
	logger.Infof("- Recent average (%d blocks): %.2fs", 
		len(n.RecentIntervals), recentAvg)
	
	n.LastBlockTime = currentTime
	n.BlockCount++
}
