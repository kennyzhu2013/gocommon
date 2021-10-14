/*
@Time : 2019/7/19 10:16
@Author : kenny zhu
@File : cache
@Software: GoLand
@Others:
*/
package apollo_config

import (
	"common/log/log"
	"github.com/coocood/freecache"
	"strconv"
	"sync"

	. "common/apollo-config/base"
)

const (
	empty = ""

	// 50m
	apolloConfigCacheSize = 50 * 1024 * 1024

	// 2 minute
	configCacheExpireTime = 120
)

var (
	currentConnApolloConfig = &currentApolloConfig{}

	// config from apollo
	apolloConfigCache = freecache.NewCache(apolloConfigCacheSize)
)

type currentApolloConfig struct {
	l      sync.RWMutex
	config *ApolloConnConfig
}

func updateApolloConfig(apolloConfig *ApolloConfig, isBackupConfig bool) {
	if apolloConfig == nil {
		log.Error("apolloConfig is null,can't update!")
		return
	}
	// get change list
	changeList := updateApolloConfigCache(apolloConfig.Configurations, configCacheExpireTime)

	if len(changeList) > 0 {
		// create config change event base on change list
		event := createConfigChangeEvent(changeList, apolloConfig.NamespaceName)

		// push change event to channel
		PushChangeEvent(event)
	}

	// update apollo connection config
	currentConnApolloConfig.l.Lock()
	defer currentConnApolloConfig.l.Unlock()

	currentConnApolloConfig.config = &apolloConfig.ApolloConnConfig

	if isBackupConfig{
		// write config file async
		go writeConfigFile(apolloConfig, appConfig.getBackupConfigPath())
	}
}

func updateApolloConfigCache(configurations map[string]string, expireTime int) map[string]*ConfigChange {
	if (configurations == nil || len(configurations) == 0) && apolloConfigCache.EntryCount() == 0 {
		return nil
	}

	// get old keys
	mp := map[string]bool{}
	it := apolloConfigCache.NewIterator()
	for en := it.Next(); en != nil; en = it.Next() {
		mp[string(en.Key)] = true
	}

	changes := make(map[string]*ConfigChange)

	if configurations != nil {
		// update new
		// keys
		for key, value := range configurations {
			// key state insert or update
			// insert
			if !mp[key] {
				changes[key] = CreateAddConfigChange(value)
			} else {
				// update
				oldValue, _ := apolloConfigCache.Get([]byte(key))
				if string(oldValue) != value {
					changes[key] = CreateModifyConfigChange(string(oldValue), value)
				}
			}

			_ = apolloConfigCache.Set([]byte(key), []byte(value), expireTime)
			delete(mp, string(key))
		}
	}

	// remove del keys
	for key := range mp {
		// get old value and del
		oldValue, _ := apolloConfigCache.Get([]byte(key))
		changes[key] = CreateDeletedConfigChange(string(oldValue))

		apolloConfigCache.Del([]byte(key))
	}

	return changes
}

// base on changeList create Change event
func createConfigChangeEvent(changes map[string]*ConfigChange, nameSpace string) *ChangeEvent {
	return &ChangeEvent{
		Namespace: nameSpace,
		Changes:   changes,
	}
}

func touchApolloConfigCache() error {
	updateApolloConfigCacheTime(configCacheExpireTime)
	return nil
}

func updateApolloConfigCacheTime(expireTime int) {
	it := apolloConfigCache.NewIterator()
	for i := int64(0); i < apolloConfigCache.EntryCount(); i++ {
		entry := it.Next()
		if entry == nil {
			break
		}
		_ = apolloConfigCache.Set([]byte(entry.Key), []byte(entry.Value), expireTime)
	}
}

func GetApolloConfigCache() *freecache.Cache {
	return apolloConfigCache
}

func GetCurrentApolloConfig() *ApolloConnConfig {
	currentConnApolloConfig.l.RLock()
	defer currentConnApolloConfig.l.RUnlock()

	return currentConnApolloConfig.config

}

func getConfigValue(key string) interface{} {
	value, err := apolloConfigCache.Get([]byte(key))
	if err != nil {
		log.Errorf("get config value fail! key:%s, err:%s", key, err)
		return empty
	}

	return string(value)
}

func getValue(key string) string {
	value := getConfigValue(key)
	if value == nil {
		return empty
	}

	return value.(string)
}

func GetStringValue(key string, defaultValue string) string {
	value := getValue(key)
	if value == empty {
		return defaultValue
	}

	return value
}

func GetIntValue(key string, defaultValue int) int {
	value := getValue(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Debug("convert to int fail!error:", err)
		return defaultValue
	}

	return i
}

func GetFloatValue(key string, defaultValue float64) float64 {
	value := getValue(key)

	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Debug("convert to float fail!error:", err)
		return defaultValue
	}

	return i
}

func GetBoolValue(key string, defaultValue bool) bool {
	value := getValue(key)

	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Debug("convert to bool fail!error:", err)
		return defaultValue
	}

	return b
}