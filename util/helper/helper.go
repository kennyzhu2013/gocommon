/*
@Time : 2019/7/3 16:58
@Author : kenny zhu
@File : helper
@Software: GoLand
@Others:
*/
package helper

import "strconv"

// 0-9:
var cipherTab = map[string]string { "acd":"0", "bfg":"1", "cjk":"2", "dop":"3", "eqz":"4", "fyt":"5", "gmx":"6", "hvn":"7", "jpw":"8", "kri":"9"}

// get timestamp from session id
func GetTimeStampFromSessionId(sessionId string)  int64 {
	const StartStamp int64 = 1561690135700
	const TimeStampLeft = 21

	// session,_ := strconv.Atoi( sessionId )
	session,_ := strconv.ParseInt(sessionId, 10, 64)
	return ( session >> TimeStampLeft) * 100 + StartStamp
}

func DecodeCompanyId(source string) string {
	var dest string
	for i:=0; i<len(source);i+=3 {
		key := source[i:i+3]
		dest += cipherTab[key]
	}
	return dest
}