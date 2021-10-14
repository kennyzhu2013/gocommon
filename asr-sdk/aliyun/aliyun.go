/*
@Time : 2020/4/24 11:52
@Author : kenny zhu
@File : aliyun
@Software: GoLand
@Others:
*/
package aliyun

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

//常量
const (
	SentenceBegin              = "SentenceBegin"              //检测到了一句话的开始
	SentenceEnd                = "SentenceEnd"                //检测到了一句话的结束，并附带返回该句话的识别结果
	TranscriptionResultChanged = "TranscriptionResultChanged" //表示句子的中间识别结果,识别结果发生了变化
	TranscriptionStarted       = "TranscriptionStarted"       ////step1: start开始事件
	TranscriptionCompleted     = "TranscriptionCompleted"     ////step3:end结束事件

	//阿里云的配置信息
	AccessKeyId     = "LTAI4G8gXRu54GJf7Th2gqXa"
	AccessKeySecret = "tBf8jfiiVHiOeMP1GOJFKXgGcof3zY"
	AppKey          = "fKRJCYA3nfs3RK2Z"

	//连接的地址
	Address = "nls-gateway.cn-shanghai.aliyuncs.com"
	AddPath = "/ws/v1"

	//音频的一些信息
	SampleRate16 = 16000
	SampleRate8  = 8000
	FormatPCM    = "pcm"
)

/**
  回调接口
*/
type CallBack func(taskId string, index int32, result string)

/**
  阿里云的封装
*/
type AliYunWrapper struct {
	TokenInfoCache TokenResp       //阿里云的token
	Conn           *websocket.Conn //ws长链
	AsrCallBack    CallBack        //接口回调
	TaskId         string          //taskId
}

/**
阿里云登录的token信息
*/
type TokenInfo struct {
	ExpireTime int64
	Id         string
	UserId     string
}

/**
  token的返回resp
*/
type TokenResp struct {
	NlsRequestId string
	RequestId    string
	Token        TokenInfo
}

/**
connect aliyun
step 1：start and confirm
*/
type StartSpeechTranscriber struct {
	Header  Header        `json:"header,omitempty"`
	Context HeaderContent `json:"context,omitempty"`
	Payload ReqPayload    `json:"payload,omitempty"`
}

/**
  step3 stop and complete
*/
type EndSpeechTranscriber struct {
	Header Header `json:"header,omitempty"`
}

/**
请求的 payload
*/
type ReqPayload struct {
	EnableIgnoreSentenceTimeout    bool   `json:"enable_ignore_sentence_timeout,omitempty"`       //设置是否忽略实时识别中的单句识别超时，默认是False
	SampleRate                     int32  `json:"sample_rate,omitempty"`                          //音频采样率，默认是16000Hz，请根据音频的采样率在管控台对应项目中配置支持该采样率及场景的模型
	Format                         string `json:"format,omitempty,omitempty"`                     //音频编码格式，默认是pcm(无压缩的pcm文件或wav文件)，16bit采样位数的单声道
	EnableIntermediateResult       bool   `json:"enable_intermediate_result,omitempty,omitempty"` //是否返回中间识别结果，默认是False
	EnableInverseTextNormalization bool   `json:"enable_inverse_text_normalization,omitempty"`    //是否在后处理中执行ITN，设置为true时，中文数字将转为阿拉伯数字输出，默认是False 。注意：不会对词信息进行ITN转换。
	EnablePunctuationPrediction    bool   `json:"enable_inverse_text_normalization,omitempty"`    //是否在后处理中添加标点，默认是False
}

/**
请求的头部
*/
type Header struct {
	Namespace string `json:"namespace,omitempty"`  //StartSpeechTranscriber
	Name      string `json:"name,omitempty"`       //StartTranscription
	TaskId    string `json:"task_id,omitempty"`    //任务全局唯一ID，请记录该值，便于排查问题
	MessageId string `json:"message_id,omitempty"` //本次消息的ID
	AppKey    string `json:"appkey,omitempty"`     //控台创建的项目Appkey
}

/**
  start的请求的头部的内容
*/
type HeaderContent struct {
	Sdk     HeaderContentSdK     `json:"sdk,omitempty"`
	Network HeaderContentNetwork `json:"network,omitempty"`
}

/**
  start的请求的头部的sdk信息
*/
type HeaderContentSdK struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

/**
  start的请求的头部的网络信息
*/
type HeaderContentNetwork struct {
	UpgradeCost int32 `json:"upgrade_cost"`
	ConnectCost int32 `json:"connect_cost"`
}

//返回接口，json数据解析
type SpeechRespHeader struct {
	Namespace  string `json:"namespace"`   //消息所属的命名空间
	Name       string `json:"name"`        //消息名称，SentenceBegin表示一个句子的开始
	Status     int    `json:"status"`      //状态码，表示请求是否成功，见服务状态码
	StatusText string `json:"status_text"` //状态消息
	TaskId     string `json:"task_id"`     //任务全局唯一ID，请记录该值，便于排查问题
	MessageId  string `json:"message_id"`  //本次消息的ID
}

/**
start ,stop 返回的结构体
*/
type SpeechResp struct {
	Header SpeechRespHeader `json:"header"`
}

/**
step 2：send and recognize
 asr返回的结构体的word
*/
type Word struct {
	Text      string `json:"text"`      //文本
	StartTime int64  `json:"startTime"` //词开始时间，单位为毫秒
	EndTime   int64  `json:"endTime"`   //词结束时间，单位为毫秒
}

/**
step 2：send and recognize
 asr返回的结构体的playload
*/
type ResponsePayLoad struct {
	Index      int32   `json:"index"`      //句子编号，从1开始递增
	Time       int64   `json:"time"`       //当前已处理的音频时长，单位是毫秒
	Result     string  `json:"result"`     //当前句子的识别结果
	Words      []Word  `json:"words"`      //当前句子的词信息，需要将enable_words设置为true
	Confidence float32 `json:"confidence"` //当前句子识别结果的置信度，取值范围[0.0, 1.0]，值越大表示置信度越高
}

/**
step 2：send and recognize
 asr返回的结构体
*/
type SpeechTranscriberResponse struct {
	Header  SpeechRespHeader `json:"header"`  //
	PayLoad ResponsePayLoad  `json:"payload"` //
}

type AsrResult struct {
	Result    string //asr的结果
	Index     int32  //asr的结果下标
	EventName string //事件名称
	Time      int64  //asr结束的时间
}

/**
获取阿里云的token信息
*/
func (aliYun *AliYunWrapper) getToken() TokenResp {
	//这里可以内存缓存下 不用次次去请求，不过期不请求
	tokenCache := aliYun.TokenInfoCache.Token
	if tokenCache.Id != "" && !aliYun.isExpire(tokenCache.ExpireTime) {
		log.Println(" use cache token info  ")
		return aliYun.TokenInfoCache
	}

	client, err := sdk.NewClientWithAccessKey("cn-shanghai", AccessKeyId, AccessKeySecret)
	if err != nil {
		panic(err)
	}
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Domain = "nls-meta.cn-shanghai.aliyuncs.com"
	request.ApiName = "CreateToken"
	request.Version = "2019-02-28"
	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		panic(err)
	}

	httpStatus := response.GetHttpStatus()
	log.Print(httpStatus)

	var tokenInfo TokenResp
	if httpStatus == 200 {
		result := response.GetHttpContentString()
		log.Println(result)
		json.Unmarshal([]byte(result), &tokenInfo) //json解析
		// 打印索引和其他数据
		log.Println(`NlsRequestId：`, tokenInfo.NlsRequestId, "\t")
		log.Println(`ExpireTime：`, tokenInfo.Token.ExpireTime, "\t")
		log.Println(`Id：`, tokenInfo.Token.Id, "\t")
	} else {
		statusStr := strconv.Itoa(httpStatus)
		log.Println("httpStatus != 200 , = " + statusStr)
	}
	//缓存这个值
	aliYun.TokenInfoCache = tokenInfo
	return tokenInfo
}

/**
判断token是否过时了
*/
func (aliYun *AliYunWrapper) isExpire(expireTime int64) bool {
	curTime := time.Now().Unix()
	if curTime > expireTime {
		return true
	}
	return false
}

/**
生成随机的Id
*/
func (aliYun *AliYunWrapper) GetId() string {
	id := strings.ReplaceAll(uuid.New().String(), "-", "")
	log.Println(id)
	return id
}

/**
获得start的请求的json
*/
func (aliYun *AliYunWrapper) GetStartReqJson(taskId string) string {
	sdk := &HeaderContentSdK{
		Name:    "nls-sdk-java",
		Version: "2.1.6",
	}
	network := &HeaderContentNetwork{
		ConnectCost: 561,
		UpgradeCost: 203,
	}

	context := &HeaderContent{
		Sdk:     *sdk,
		Network: *network,
	}

	messageId := aliYun.GetId()
	header := &Header{
		AppKey:    AppKey,
		MessageId: messageId,
		Name:      "StartTranscription",
		Namespace: "SpeechTranscriber",
		TaskId:    taskId,
	}

	payload := &ReqPayload{
		EnableIgnoreSentenceTimeout:    false,
		SampleRate:                     SampleRate16,
		Format:                         FormatPCM,
		EnableInverseTextNormalization: false,
		EnableIntermediateResult:       false,
		EnablePunctuationPrediction:    true,
	}

	start := &StartSpeechTranscriber{
		Header:  *header,
		Payload: *payload,
		Context: *context,
	}

	startData, _ := json.Marshal(start)
	log.Println("start-json:" + string(startData))
	return string(startData)
}

/**
  stop的json
*/
func (aliYun *AliYunWrapper) GetStopReqJson() string {
	messageId := aliYun.GetId()
	taskId := aliYun.TaskId
	header := &Header{
		AppKey:    AppKey,
		MessageId: messageId,
		Name:      "StopTranscription",
		Namespace: "SpeechTranscriber",
		TaskId:    taskId,
	}

	end := &EndSpeechTranscriber{
		Header: *header,
	}

	endData, _ := json.Marshal(end)
	log.Println("end-json:" + string(endData))
	return string(endData)
}

/**
连接和发start消息
*/
func (aliYun *AliYunWrapper) ConnectAndStart() string {
	//log.SetFlags(0)
	token := aliYun.getToken()
	accessToken := token.Token.Id
	if accessToken == "" {
		log.Println(" accessToken empty")
		return ""
	}

	u := url.URL{Scheme: "wss", Host: Address, Path: AddPath}
	log.Printf("connecting to %s", u.String())

	//连接的时候 带上token
	header := http.Header{}
	header.Set("X-NLS-Token", accessToken)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		//log.Printf(" error  %s", err)
		log.Fatal("dial:", err)
	}

	//缓存这个连接
	aliYun.Conn = c

	//缓存起来就先不删除了
	//defer conn.Close() //最后关闭

	//连接成功之后，发送start and confirm
	taskId := aliYun.GetId()
	log.Printf(" start startSend ")
	suc := aliYun.SendStartConfirm(aliYun.GetStartReqJson(taskId))
	if suc {
		aliYun.TaskId = taskId //存这个taskId
		return taskId
	}
	return ""
}

/**
  发送开始start事件和确认
*/
func (aliYun *AliYunWrapper) SendStartConfirm(json string) bool {
	aliYun.Conn.WriteMessage(websocket.TextMessage, []byte(json))
	status := aliYun.readOneMessage()
	if status == 20000000 {
		return true
	}
	return false
}

func (aliYun *AliYunWrapper) SendEndComplete() {
	if aliYun.Conn != nil {
		stopJson := aliYun.GetStopReqJson()
		aliYun.Conn.WriteMessage(websocket.TextMessage, []byte(stopJson))
	}else {
		log.Println(" sendEnd ,aliYun.Conn is nil ")
	}
}

//发送字节流
func (aliYun *AliYunWrapper) SendBinary(buf []byte) {
	//发送流
	if aliYun.Conn != nil {
		aliYun.Conn.WriteMessage(websocket.BinaryMessage, buf)
	} else {
		log.Println(" sendBinary ,aliYun.Conn is nil")
	}
}

func (aliYun *AliYunWrapper) readOneMessage() int {
	//end的结果
	resp := make(chan string, 1)
	//监听
	go func() {
		_, message, err := aliYun.Conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			resp <- "err"
		}
		log.Printf("recv: %s", message)
		resp <- string(message)
	}()

	defer close(resp)
	//接收start的返回
	speedResp := &SpeechResp{}
	select {
	case respJson := <-resp:
		log.Println(respJson)
		if respJson == "err" {
			log.Println(" start , end  err ")
		} else {
			log.Println("start , end parse start result ")
			json.Unmarshal([]byte(respJson), &speedResp) //end json解析
			return speedResp.Header.Status
		}
	case <-time.After(time.Second * 5):
		log.Println("start , end wait 5s ,timeout ")
	}
	return 0
}

/**
  监听asr的结果
*/
func (aliYun *AliYunWrapper) ReadMessage() {
	for {
		_, message, err := aliYun.Conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			//失败继续读
			break
		}
		log.Printf(" asr recv--- : %s", message)

		//解析数据
		asrResult := aliYun.parseAliMessage(message)
		log.Println(asrResult.EventName + "," + asrResult.Result)
		if asrResult.EventName == TranscriptionCompleted {
			//结束直接退出
			break
		}
	}
}

func (aliYun *AliYunWrapper) parseAliMessage(message []byte) AsrResult {
	response := &SpeechTranscriberResponse{}
	json.Unmarshal(message, &response)
	name := response.Header.Name
	taskId := response.Header.TaskId
	log.Println("taskId:" + taskId + ",response:" + name)
	result := ""
	var index int32
	var time int64
	index = 0
	time = 0
	if name == SentenceEnd { //一句话结束
		result = response.PayLoad.Result
		index = response.PayLoad.Index
		time = response.PayLoad.Time
		log.Println(" SentenceEnd :" + result)
		if aliYun != nil {
			//消息通过接口 回调出去
			aliYun.AsrCallBack(taskId, index, result)
		}
	}
	//else if name == TranscriptionCompleted {
	//	respStatus := response.Header.Status
	//	log.Println(" TranscriptionCompleted :" + strconv.Itoa(respStatus))
	//} else if name == SentenceBegin ||
	//	name == TranscriptionResultChanged ||
	//	name == TranscriptionStarted {
	//	// nothing to do
	//}

	asrResult := AsrResult{
		EventName: name,
		Result:    result,
		Index:     index,
		Time:      time,
	}
	return asrResult
}

/**
  设置回调方法
*/
func (aliYun *AliYunWrapper) SetAsrCallBack(back CallBack) {
	aliYun.AsrCallBack = back
}

/**
  外部方法
*/
func asrResultCallBack(taskId string, index int32, result string) {
	log.Println("outside asrResultCallBack ", taskId, index, result)
}

//测试入口
func test() {
	aliYunWrapper := AliYunWrapper{}

	aliYunWrapper.ConnectAndStart()

	//设置结果回调
	aliYunWrapper.SetAsrCallBack(asrResultCallBack)

	//监听
	go aliYunWrapper.ReadMessage()

	//这里读文件流
	f, err := os.Open("nls-sample-16k.wav")
	if err != nil {
		fmt.Println("read fail")
		return
	}
	//把file读取到缓冲区中
	defer f.Close()
	buf := make([]byte, 1024)

	for {
		//从file读取到buf中
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("read buf fail", err)
			return
		}
		//说明读取结束
		if n == 0 {
			break
		}
		//发送流
		aliYunWrapper.SendBinary(buf)
		//发流之后就sleep 20ms
		time.Sleep(time.Millisecond * 20)
	}
	//time.Sleep(time.Millisecond * 50)
	aliYunWrapper.SendEndComplete()
	time.Sleep(time.Millisecond * 50)
	//关闭流
	aliYunWrapper.Conn.Close()
}
