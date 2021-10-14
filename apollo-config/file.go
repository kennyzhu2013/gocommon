/*
@Time : 2019/7/18 16:37
@Author : kenny zhu
@File : file cache for apollo
@Software: GoLand
@Others:
*/
package apollo_config

import (
	"common/log/log"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	. "common/apollo-config/base"
)

const FILE = "apolloConfig.json"
var configFile=""

// write config to file
func writeConfigFile(config *ApolloConfig,configPath string)error{
	if config==nil{
		log.Error("apollo config is null can not write backup file")
		return errors.New("apollo config is null can not write backup file")
	}
	file, e := os.Create(getConfigFile(configPath))
	if e!=nil{
		log.Errorf("writeConfigFile fail,error:",e)
		return e
	}
	defer  file.Close()

	return json.NewEncoder(file).Encode(config)
}

// get real config file
func getConfigFile(configDir string) string {
	if configFile == "" {
		if configDir!="" {
			configFile=fmt.Sprintf("%s/%s",configDir,FILE)
		}else{
			configFile=FILE
		}

	}
	return configFile
}

// load config from file
func loadConfigFile(configDir string) (*ApolloConfig,error){
	configFilePath := getConfigFile(configDir)
	log.Info("load config file from :",configFilePath)
	file, e := os.Open(configFilePath)
	if e != nil {
		log.Error("loadConfigFile fail,error:", e.Error())
		return nil,e
	}
	defer file.Close()
	config := &ApolloConfig{}
	e = json.NewDecoder(file).Decode(config)

	if e != nil{
		log.Error("loadConfigFile fail,error: ", e.Error())
		return nil,e
	}

	return config,e
}