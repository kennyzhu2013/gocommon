/*
@Time : 2019/7/18 16:48
@Author : kenny zhu
@File : notify
@Software: GoLand
@Others:
*/
package apollo_config

import (
	"common/log/log"
	"encoding/json"
	"errors"
	"sync"
	"time"

	. "common/apollo-config/base"
)

const(
	defaultNotificationId = -1
)

var(
	allNotifications *notificationsMap
)

type notification struct {
	NamespaceName string `json:"namespaceName"`
	NotificationId int64 `json:"notificationId"`
}

type notificationsMap struct {
	notifications map[string]int64
	sync.RWMutex
}


type apolloNotify struct {
	NotificationId int64 `json:"notificationId"`
	NamespaceName string `json:"namespaceName"`
}

func (n *notificationsMap) setNotify(namespaceName string,notificationId int64) {
	n.Lock()
	defer n.Unlock()
	n.notifications[namespaceName]=notificationId
}
func (n *notificationsMap) getNotifies() string {
	n.RLock()
	defer n.RUnlock()

	notificationArr:=make([]*notification,0)
	for namespaceName,notificationId:=range n.notifications{
		notificationArr=append(notificationArr,
			&notification{
				NamespaceName:namespaceName,
				NotificationId:notificationId,
			})
	}

	j,err := json.Marshal(notificationArr)

	if err!=nil{
		return ""
	}

	return string(j)
}

func initAllNotifications()  {
	appConfig := GetAppConfig(nil)

	if appConfig != nil {
		allNotifications = &notificationsMap{
			notifications: make(map[string]int64, 1),
		}

		allNotifications.setNotify(appConfig.NamespaceName, defaultNotificationId)
	}
}

type NotifyConfigComponent struct {
}

func (c *NotifyConfigComponent) Start()  {
	t2 := time.NewTimer(long_poll_interval)
	// long poll for sync
	for {
		select {
		case <-t2.C:
			_ = notifySyncConfigServices()
			t2.Reset(long_poll_interval)
		}
	}
}

func notifySyncConfigServices() error {
	remoteConfigs,err := notifyRemoteConfig(nil)

	if err!=nil||len(remoteConfigs)==0{
		return err
	}

	updateAllNotifications(remoteConfigs)

	// sync all config
	_ = autoSyncConfigServices(nil)

	return nil
}

func toApolloConfig(resBody []byte) ([]*apolloNotify,error) {
	remoteConfig:=make([]*apolloNotify,0)

	err:=json.Unmarshal(resBody,&remoteConfig)

	if err!=nil{
		log.Error("Unmarshal Msg Fail,Error:",err.Error())
		return nil,err
	}
	return remoteConfig,nil
}

func notifyRemoteConfig(newAppConfig *AppConfig) ([]*apolloNotify,error) {
	appConfig := GetAppConfig(newAppConfig)
	if appConfig == nil{
		log.Error("can not find apollo config!please confirm!")
		return nil, errors.New( "can not find apollo config!please confirm! " )
	}
	urlSuffix := getNotifyUrlSuffix(allNotifications.getNotifies(), appConfig,newAppConfig)

	// log.Debugf("allNotifications.getNotifies():%s",allNotifications.getNotifies())

	notifies ,err := requestRecovery(appConfig,&ConnectConfig{
		Uri:urlSuffix,
		Timeout:nofity_connect_timeout,
	},&CallBack{
		SuccessCallBack: func(responseBody []byte) (interface{}, error) {
			return toApolloConfig(responseBody)
		},
		NotModifyCallBack: touchApolloConfigCache,
	})

	if notifies==nil{
		return nil,err
	}

	return notifies.([]*apolloNotify),err
}

func updateAllNotifications(remoteConfigs []*apolloNotify) {
	for _,remoteConfig:=range remoteConfigs{
		if remoteConfig.NamespaceName==""{
			continue
		}

		allNotifications.setNotify(remoteConfig.NamespaceName,remoteConfig.NotificationId)
	}
}

func autoSyncConfigServicesSuccessCallBack(responseBody []byte)(o interface{},err error){
	apolloConfig,err := CreateApolloConfigWithJson(responseBody)

	if err!=nil{
		log.Error("Unmarshal Msg Fail,Error:", err)
		return nil,err
	}

	updateApolloConfig(apolloConfig,true)

	return nil,nil
}

func autoSyncConfigServices(newAppConfig *AppConfig) error {
	appConfig := GetAppConfig(newAppConfig)
	if appConfig == nil{
		log.Error("can not find apollo config!please confirm!")
		return errors.New( "can not find apollo config!please confirm! " )
	}

	urlSuffix:=getConfigUrlSuffix(appConfig, newAppConfig)

	_,err:=requestRecovery(appConfig,&ConnectConfig{
		Uri:urlSuffix,
	},&CallBack{
		SuccessCallBack:autoSyncConfigServicesSuccessCallBack,
		NotModifyCallBack:touchApolloConfigCache,
	})

	return err
}