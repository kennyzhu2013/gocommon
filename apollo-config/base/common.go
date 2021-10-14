/*
@Time : 2019/7/18 16:47
@Author : kenny zhu
@File : common
@Software: GoLand
@Others:
*/
package base

import (
	"encoding/json"
	"sync"
)

type AbsComponent interface {
	Start()
}

func StartRefreshConfig(component AbsComponent)  {
	component.Start()
}

type ApolloConnConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	ReleaseKey string `json:"releaseKey"`
	sync.RWMutex
}

type ApolloConfig struct {
	ApolloConnConfig
	Configurations map[string]string `json:"configurations"`
}

func CreateApolloConfigWithJson(b []byte) (*ApolloConfig,error) {
	apolloConfig:=&ApolloConfig{}
	err:=json.Unmarshal(b,apolloConfig)
	if err != nil {
		return nil,err
	}
	return apolloConfig,nil
}
