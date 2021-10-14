/*
@Time : 2019/4/23 15:48 
@Author : kenny zhu
@File : common.go
@Software: GoLand
@Others:
*/
package monitor

import (
	"common/registry"
	"time"
)


type Heartbeat struct {
	id string// The ID of the heartbeat. Will generally be the Node ID.
	service *registry.Service // The service sending the heartbeat, support multi-service
	timestamp int64 // Unix time at which this was sent
	ttl int64  // The time to live for this heartbeat,  for out-dated.
	weights int // node weight.
}

const (
	// default ttl value
    HeartBeatTTL = time.Second * 60
	HeartBeatCheck = time.Second * 15 //TODO: 会影响灰度发布统计精度..

	// state define here.
	ServiceStatus = "ServiceStatus"
	NormalState = "normal"
	UpThresh = "upThresh"
	DownThresh = "downThresh"
	DeleteState = "deleted"
)

