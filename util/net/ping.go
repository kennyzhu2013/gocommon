/*
@Time : 2019/6/10 11:57
@Author : kenny zhu
@File : ping
@Software: GoLand
@Others:
*/
package net

import (
	"common/log/log"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type PingStat struct {
	Ip     string
	SendN  int
	LostN  int
	RecvN  int
	ShortT int
	LongT  int
	SumT   int
}

// ping协程函数.
func Ping(host string, args map[string]interface{}) PingStat {
	var count int
	var size int
	var timeout int64
	var neverstop bool
	count = args["n"].(int)
	size = args["l"].(int)
	timeout = args["w"].(int64)
	neverstop = args["t"].(bool)

	cname, _ := net.LookupCNAME(host)
	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	ip := conn.RemoteAddr()
	log.Info("正在 Ping " + cname + " [" + ip.String() + "] 具有 32 字节的数据:")

	var seq int16 = 1
	id0, id1 := genidentifier(host)
	const ECHO_REQUEST_HEAD_LEN = 8

	sendN := 0
	recvN := 0
	lostN := 0
	shortT := -1
	longT := -1
	sumT := 0

	for count > 0 || neverstop {
		sendN++
		var msg []byte = make([]byte, size+ECHO_REQUEST_HEAD_LEN)
		msg[0] = 8                        // echo
		msg[1] = 0                        // code 0
		msg[2] = 0                        // checksum
		msg[3] = 0                        // checksum
		msg[4], msg[5] = id0, id1         // identifier[0] identifier[1]
		msg[6], msg[7] = gensequence(seq) // sequence[0], sequence[1]

		length := size + ECHO_REQUEST_HEAD_LEN

		check := checkSum(msg[0:length])
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err = net.DialTimeout("ip:icmp", host, time.Duration(timeout*1000*1000))

		if err != nil {
			seq++
			count--
			checkError(err)
			return PingStat{Ip: ip.String(), SendN: sendN, LostN: sendN, RecvN: 0, ShortT: -1, LongT: -1, SumT: -1}
		}

		starttime = time.Now()
		conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const ECHO_REPLY_HEAD_LEN = 20

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		_ = n

		var endduration int = int(int64(time.Since(starttime)) / (1000 * 1000))

		sumT += endduration

		time.Sleep(1000 * 1000 * 1000)

		if err != nil || receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] || receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] || receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
			lostN++
			// fmt.Println("对 " + cname + "[" + ip.String() + "]" + " 的请求超时。")
		} else {
			if shortT == -1 {
				shortT = endduration
			} else if shortT > endduration {
				shortT = endduration
			}
			if longT == -1 {
				longT = endduration
			} else if longT < endduration {
				longT = endduration
			}
			recvN++
			ttl := int(receive[8])
			//			fmt.Println(ttl)
			log.Info("来自 " + cname + "[" + ip.String() + "]" + " 的回复: 字节=32 时间=" + strconv.Itoa(endduration) + "ms TTL=" + strconv.Itoa(ttl))
		}

		seq++
		count--
	}

	// all result..
	// stat(ip.String(), sendN, lostN, recvN, shortT, longT, sumT)
	// c <- 1
	return PingStat{Ip: ip.String(), SendN: sendN, LostN: lostN, RecvN: recvN, ShortT: shortT, LongT: longT, SumT: sumT}
}

func checkSum(msg []byte) uint16 {
	sum := 0

	length := len(msg)
	for i := 0; i < length-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if length%2 == 1 {
		sum += int(msg[length-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		// os.Exit(1)
	}
}

func gensequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genidentifier(host string) (byte, byte) {
	return host[0], host[1]
}

func Stat(result PingStat) {
	log.Info("\n" + result.Ip, " 的 Ping 统计信息:")
	log.Info("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%d%% 丢失)，\n", result.SendN, result.RecvN, result.LostN, int(result.LostN*100/result.SendN))
	log.Info("往返行程的估计时间(以毫秒为单位):")
	if result.RecvN != 0 {
		log.Info("    最短 = %dms，最长 = %dms，平均 = %dms\n", result.ShortT, result.LongT, result.SumT/result.SendN)
	}
}
