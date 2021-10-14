/*
@Time : 2019/7/18 15:12
@Author : kenny zhu
@File : appconfig
@Software: GoLand
@Others:
*/
package apollo_config

import (
	"strings"
	"time"
)

var (
	long_poll_interval = 5 * time.Second // 2s
	long_poll_connect_timeout  = 1 * time.Minute // 1m

	connect_timeout  = 1 * time.Second //1s
	// notify timeout
	nofity_connect_timeout  = 10 * time.Minute // 10m
	// for on error retry
	on_error_retry_interval = 1 * time.Second // 1s
	// for typed config cache of parser result, e.g. integer, double, long, etc.
	// max_config_cache_size    = 500             //500 cache key
	// config_cache_expire_time = 1 * time.Minute //1 minute

	// max retries connect apollo
	max_retries = 5

	// refresh ip list
	refresh_ip_list_interval = 20 * time.Minute // 20m

	// app config
	appConfig *AppConfig

	// real servers ip
	servers map[string]*serverInfo=make(map[string]*serverInfo,0)

	// next try connect period - 60 second
	next_try_connect_period int64=60

)

type AppConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	Ip string `json:"ip"`
	NextTryConnTime int64 `json:"-"`
	BackupConfigPath string `json:"backupConfigPath"`
}


func (c *AppConfig) getBackupConfigPath() string{
	return c.BackupConfigPath
}


func (c *AppConfig) getHost() string{
	if strings.HasPrefix(c.Ip,"http"){
		if !strings.HasSuffix(c.Ip,"/") {
			return c.Ip+"/"
		}
		return c.Ip
	}
	return "http://"+c.Ip+"/"
}

// if this connect is fail will set this time
func (c *AppConfig) setNextTryConnTime(nextTryConnectPeriod int64){
	c.NextTryConnTime=time.Now().Unix()+nextTryConnectPeriod
}

// is connect by ip directly
// false : no
// true : yes
func (c *AppConfig) isConnectDirectly() bool{
	if c.NextTryConnTime>=0&& c.NextTryConnTime>time.Now().Unix(){
		return true
	}

	return false
}

func (c *AppConfig) selectHost() string{
	if !c.isConnectDirectly(){
		return c.getHost()
	}

	for host,server:=range servers{
		// if some node has down then select next node
		if server.IsDown{
			continue
		}
		return host
	}

	return ""
}

type serverInfo struct {
	AppName string `json:"appName"`
	InstanceId string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown bool `json:"-"`
}


func setDownNode(host string) {
	if host=="" || appConfig==nil{
		return
	}

	if host==appConfig.getHost(){
		appConfig.setNextTryConnTime(next_try_connect_period)
	}

	for key,server:=range servers{
		if key==host{
			server.IsDown=true
			break
		}
	}
}

func GetAppConfig(newAppConfig *AppConfig) *AppConfig  {
	if newAppConfig !=nil{
		return newAppConfig
	}
	return appConfig
}

