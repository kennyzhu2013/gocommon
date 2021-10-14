/*
@Time : 2019/7/18 12:05
@Author : kenny zhu
@File : apollo
@Software: GoLand
@Others:
*/
package apollo_config

import (
	"common/log/log"
	"common/util/addr"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	. "common/apollo-config/base"
)

const (
	appConfigFileName  = "app.properties"
	defaultCluster = "default"
	defaultNamespace = "application"

)

func loadJsonConfig(fileName string) error{
	fs, err := ioutil.ReadFile(fileName)
	if err != nil {
		return errors.New("Fail to read config file:" + err.Error())
	}

	appConfig = &AppConfig{
		Cluster: defaultCluster,
		NamespaceName:defaultNamespace,
	}
	err = json.Unmarshal(fs,appConfig)
	if err != nil{
		return errors.New("Load Json Config fail:" + err.Error())
	}

	return nil
}

func initCache()  {
	// updateApolloConfig
	apolloConfig := &ApolloConfig{}
	apolloConfig.AppId = appConfig.AppId
	apolloConfig.Cluster = appConfig.Cluster
	apolloConfig.NamespaceName = appConfig.NamespaceName
	updateApolloConfig(apolloConfig,false)
}

// init from default file
func InitAndStartApollo() error {
	err := loadJsonConfig(appConfigFileName)
	if err != nil {
		return err
	}

	initCache()
	return StartApolloClient(appConfig)
}

// must start watch config
func StartApolloClient(config *AppConfig) error {
	if appConfig == nil {
		appConfig = config
		initCache()
	}
	initAllNotifications()

	// init server ip list
	go initServerIpList()

	// first sync
	err := notifySyncConfigServices()

	// first sync fail then load config file
	if err != nil{
		config, _ := loadConfigFile(appConfig.BackupConfigPath)
		if config != nil{
			// update cache
			updateApolloConfig(config,false)
		}
	}

	// start long poll sync config
	go StartRefreshConfig( &NotifyConfigComponent{} )
	log.Info("apollo config start finished , error:", err)

	return nil
}

// set timer for update ip list
// interval : 20 seconds
func initServerIpList() {
	_ = syncServerIpList(nil)

	t2 := time.NewTimer(refresh_ip_list_interval)
	for {
		select {
		case <-t2.C:
			_ = syncServerIpList(nil)
			t2.Reset(refresh_ip_list_interval)
		}
	}
}

// sync ip list from server
// then
// 1.update cache
// 2.store in disk
func syncServerIpList(newAppConfig *AppConfig) error{
	if appConfig == nil{
		log.Error("can not find apollo config!please confirm!")
		return errors.New("can not find apollo config!please confirm! ")
	}

	_,err := request(getServicesConfigUrl(appConfig),&ConnectConfig{
	},&CallBack{
		SuccessCallBack:syncServerIpListSuccessCallBack,
	})

	return err
}

// 分布式部署情况下获取ip列表..
func syncServerIpListSuccessCallBack(responseBody []byte)(o interface{},err error){
	log.Debug("get all server info:",string(responseBody))

	tmpServerInfo:=make([]*serverInfo,0)

	err= json.Unmarshal(responseBody,&tmpServerInfo)

	if err!=nil{
		log.Error("Unmarshal json Fail,Error:",err.Error())
		return
	}

	if len(tmpServerInfo) == 0 {
		log.Info("get no real server!")
		return
	}

	for _,server :=range tmpServerInfo {
		if server==nil{
			continue
		}
		servers[server.HomepageUrl]=server
	}
	return
}

/*******************************************url create functions****************************************************/
func getNotifyUrlSuffix(notifications string,config *AppConfig,newConfig *AppConfig) string{
	if newConfig!=nil{
		return ""
	}
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}

func getServicesConfigUrl(config *AppConfig) string{
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.getHost(),
		url.QueryEscape(config.AppId),
		addr.GetInternal())
}

func getConfigUrlSuffix(config *AppConfig, newConfig *AppConfig) string{
	if newConfig != nil{
		return ""
	}
	current := GetCurrentApolloConfig()
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(config.NamespaceName),
		url.QueryEscape(current.ReleaseKey),
		addr.GetInternal())
}