/*
@Time : 2019/6/19 14:15
@Author : kenny zhu
@File : strategy
@Software: GoLand
@Others:
*/
package selector

import (
	"common/log/log"
	"common/monitor"
	"common/registry"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random is a random strategy algorithm for node selection
func Random(services []*registry.Service) Next {
	var nodes []*registry.Node

	for _, service := range services {
		nodes = append(nodes, service.Nodes...)
	}

	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, ErrNoneAvailable
		}

		if len(nodes) == 1 {
			return nodes[0], nil
		}

		i := rand.Int() % len(nodes)
		return nodes[i], nil
	}
}


// RoundRobin is a round robin strategy algorithm for node selection,
// WARN : Here not supports go-routines.
func RoundRobin(services []*registry.Service) Next {
	var nodes []*registry.Node

	for _, service := range services {
		nodes = append(nodes, service.Nodes...)
	}

	var i = rand.Int()
	var mtx sync.Mutex

	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, ErrNoneAvailable
		}

		mtx.Lock()
		node := nodes[i%len(nodes)]
		i++
		mtx.Unlock()

		return node, nil
	}
}


// support go routines
var (
	selectKey uint = 0
	mtx sync.Mutex
)

//  use round select..
// server information must transfer here.
func RoundBinSelect(services []*registry.Service) Next {
	if len(services) == 0 {
		return func() (*registry.Node, error) {
			return nil, ErrNotFound
		}
	}

	// flatten
	var nodes []*registry.Node = nil
	var delNodes []*registry.Node = nil

	// Filter the nodes for serverTag marked by the server..
	for _, service := range services {
		for _, node := range service.Nodes {
			//if node.Metadata["serverTag"] == DefaultRouter.ServerTag {
			if bHealthNodesByWeights(node) {
				nodes = append(nodes, node)
			} else if bDeletedNodes(node) {
				delNodes = append(delNodes, node)
			} else {
				// de register node to fix?..
			}
			//}
		}
	}

	if len(nodes) == 0 {
		if len(delNodes) == 0 {
			return func() (*registry.Node, error) {
				return nil, ErrNotFound
			}
		} else {
			// no other nodes, select delete nodes.
			nodes = delNodes
		}
	}

	// Round bin..
	return func() (*registry.Node, error) {
		mtx.Lock()
		defer mtx.Unlock()
		selectKey++
		log.Infof("roundBinSelect select key is:%v", selectKey)
		return nodes[ selectKey % uint(len(nodes)) ], nil
	}
}

// add stats info to select.
// weights is set by clients.
// support 70%-90%, Bucket algorithm
func bHealthNodesByWeights(node *registry.Node) bool {
	// filter weights data.
	nodeMetas := node.Metadata
	timeNow := time.Now().Unix()
	if ts, ok := nodeMetas["timestamp"]; ok {
		timestamp,_ := strconv.ParseInt(ts, 10, 64)
		if timestamp + int64( monitor.HeartBeatTTL.Seconds() ) < timeNow {
			// long time no heartbeats
			return false
		}

	}

	if status, ok := nodeMetas[monitor.ServiceStatus]; ok {
		if status != monitor.NormalState{
			return false
		}
	}

	// default
	return true
}

func bDeletedNodes(node *registry.Node) bool {
	// filter weights data.
	nodeMetas := node.Metadata
	if status, ok := nodeMetas[monitor.ServiceStatus]; ok {
		if status == monitor.DeleteState {
			return true
		}
	}

	// default
	return false
}