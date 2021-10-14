/*
* Copyright(C),2019-2029, email: 277251257@qq.com
* Author:  kennyzhu
* Version: 1.0.0
* Date:    2021/4/8 14:35
* Description:
 */
package process

import (
	log "common/log/newlog"
	logDebug "common/util/debug"
	"fmt"
	"reflect"
	"runtime"
)

type PanicReportCallBack func(req *PanicReq) interface{}

type PanicReport struct {
	service   string
	host      string
	lastPanic float64
	callback  PanicReportCallBack
}

type PanicReq struct {
	Service   string `json:"service"`
	ErrorInfo string `json:"error_info"`
	Stack     string `json:"stack"`
	LogId     string `json:"log_id"`
	FuncName  string `json:"func_name"`
	Host      string `json:"host"`
	CallId    string `json:"call_id"`
}

var DefaultPanicReport *PanicReport

func InitPanicReport(serviceName, hostName string, callback PanicReportCallBack) {
	DefaultPanicReport = &PanicReport{service: serviceName, host: hostName, callback: callback}
	/*DefaultPanicReport.callback = func() float64 {
		panicCount := DefaultPanicReport.lastPanic
		if DefaultPanicReport.lastPanic > 10 {
			DefaultPanicReport.lastPanic = 0
		}
		return panicCount
	}*/
}

func DeferLocalPanicFunc(id, funcName string, args ...interface{}) {
	if v := recover(); v != nil {
		log.Errorf("[%v] [%v] panic: %v, args: [%v]", id, funcName, v, args)
		logDebug.LogLocalStacks()
	}
}

func (r *PanicReport) ReportPanic(id, funcName, errInfo, stack string) {
	r.callback(&PanicReq{
		Service:   r.service,
		Host:      r.host,
		ErrorInfo: errInfo,
		Stack:     stack,
		FuncName:  funcName,
		CallId:    id,
	})
}

func (r *PanicReport) RecoverFromPanic(id, funcName string, err interface{}) {
	buf := make([]byte, 64<<10)
	buf = buf[:runtime.Stack(buf, false)]
	if len(id) == 0 {
		id = "-"
	}
	log.Warnf("[%v] [%v] panic: %v, stack: %s", id, funcName, err, string(buf))
	// Async handle callback
	if r.callback != nil {
		go r.ReportPanic(id, funcName, fmt.Errorf("%v", err).Error(), string(buf))
	}
	return
}

func (r *PanicReport) GetPanicReportCallBack() PanicReportCallBack {
	return r.callback
}

func (r *PanicReport) SetPanicReportCallBack(callback PanicReportCallBack) {
	r.callback = callback
}

func SafeExecuteFunc(id string, exec func()) {
	defer func() {
		if v := recover(); v != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(exec).Pointer()).Name()
			DefaultPanicReport.RecoverFromPanic(id, funcName, v)
		}
	}()

	exec()
}

func SafeExecuteFuncWithName(id, funcName string, exec func()) {
	defer func() {
		if v := recover(); v != nil {
			DefaultPanicReport.RecoverFromPanic(id, funcName, v)
		}
	}()

	exec()
}
