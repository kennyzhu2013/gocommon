/*
@Time : 2019/8/1 16:19
@Author : kenny zhu
@File : client
@Software: GoLand
@Others:
*/
package main

import (
	"common/util/addr"
	"fmt"
	"time"
	"unsafe"

	netlink "common/netlink-client"
)

// net link family
const NETLINK_USER = 22
const USER_MSG = NETLINK_USER + 1
const USER_PORT = 150

//4+4+2+2+2 = 14
var entry struct{
	lAddr uint32
	rAddr uint32
	lport uint16
	rPort uint16
	localport uint16
}

func main()  {
	entry.localport = 15000
	entry.rPort = 4100
	entry.lport = 5000
	config := &netlink.Config{ Groups: 0, Pid: uint32(entry.localport) } // USER_PORT
	conn, err := netlink.Dial(USER_MSG, config)
	if conn == nil {
		fmt.Println("netlink.Dial failed:", err)
		return
	}
	defer conn.Close()
	rwd := time.Now().Add(1 * time.Millisecond * 100)
	if err = conn.SetDeadline(rwd); err != nil {
		fmt.Println("conn.SetDeadline failed: ", err)
	}

	// construct message.
	// dataStr := "Hello test golang!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"

	// set rtp port
	entry.lAddr = uint32(addr.InetAtoN("192.168.93.128")) // 10.153.90.7
	entry.rAddr = uint32(addr.InetAtoN("192.168.93.128"))

	// memcpy
	p := unsafe.Pointer(&entry)
	q := (*[14]byte)(p)

	fmt.Printf("entry byte is: %v ", q[:])
	fmt.Printf("entry.lAddr is: %d ", entry.lAddr)
	fmt.Printf("entry.rAddr is: %d ", entry.rAddr)
	//

	// ignore length..
	m := netlink.Message{
		Header: netlink.Header{
			// Ask netlink to echo back an acknowledgement to our request.
			Flags: 0, // netlink.Request,
			Type: 1,
			PID: uint32(entry.localport), // 150, USER_PORT
			// Other fields assigned automatically by package netlink.
		},
		Data: q[:], // ;buf.Bytes(),// []byte(dataStr),
	}

	_, err = conn.Send(m)
	if err != nil {
		fmt.Println("conn.Send error:", err)
		return
	}

	replies, err := conn.Receive()
	if err != nil {
		fmt.Println("conn.Receive error:", err)
		return
	}

	// sequence number must check, ignore..
	// err := netlink.Validate(req, replies)
	// if err != nil {
	// 	fmt.Println("netlink.Validate error:", err.Error())
	// }
	if len(replies) < 1 {
		fmt.Println("conn.Receive nil, length:", len(replies))
	}

	fmt.Println("Receive response: ", string(replies[0].Data))
}
