/*
@Time : 2019/7/18 16:02
@Author : kenny zhu
@File : request
@Software: GoLand
@Others:
*/
package apollo_config

import (
	"common/log/log"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type CallBack struct {
	SuccessCallBack func([]byte)(interface{},error)
	NotModifyCallBack func()error
}

type ConnectConfig struct {
	// 设置到http.client中timeout字段
	Timeout time.Duration
	// 连接接口的uri
	Uri string
}

// http call back
func request(requestUrl string,connectionConfig *ConnectConfig,callBack *CallBack) (interface{},error){
	client := &http.Client{}
	// 如有设置自定义超时时间即使用
	if connectionConfig != nil && connectionConfig.Timeout != 0{
		client.Timeout = connectionConfig.Timeout
	} else {
		client.Timeout = connect_timeout
	}

	retry := 0
	var responseBody []byte
	var err error
	var res *http.Response
	for{
		retry++

		if retry>max_retries{
			break
		}

		res,err = client.Get(requestUrl)

		if res == nil || err != nil{
			log.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}

		// not modified break
		switch res.StatusCode {
		case http.StatusOK:
			defer res.Body.Close()
			responseBody, err = ioutil.ReadAll(res.Body)
			if err!=nil{
				log.Error("Connect Apollo Server Fail,Error:",err)
				continue
			}

			if callBack!=nil&&callBack.SuccessCallBack!=nil {
				return callBack.SuccessCallBack(responseBody)
			}else{
				return nil,nil
			}
		case http.StatusNotModified:
			log.Info("Config Not Modified:", err)
			if callBack!=nil&&callBack.NotModifyCallBack!=nil {
				return nil,callBack.NotModifyCallBack()
			}else{
				return nil,nil
			}
		default:
			log.Error("Connect Apollo Server Fail,Error:",err)
			err = errors.New("Connect Apollo Server Fail! ")
			// if error then sleep
			time.Sleep(on_error_retry_interval)
			continue
		}
	}

	log.Error("Over Max Retry Still Error,Error:",err)
	if err != nil{
		err = errors.New("Over Max Retry Still Error! ")
	}
	return nil,err
}

//
func requestRecovery(appConfig *AppConfig,
	connectConfig *ConnectConfig,
	callBack *CallBack)(interface{},error) {
	format:="%s%s"
	var err error
	var response interface{}

	// select all cluster nodes.
	for {
		host := appConfig.selectHost()
		if host == ""{
			return nil,errors.New("Try all Nodes Still Error! ")
		}

		requestUrl := fmt.Sprintf(format,host,connectConfig.Uri)
		response,err = request(requestUrl,connectConfig,callBack)
		if err == nil{
			return response,err
		}

		setDownNode(host)
	}

}