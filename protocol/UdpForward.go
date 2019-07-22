package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

var UdpForwardPort = "5927"

var UdpConnMapClient = make(map[string]net.UDPAddr)
var UdpConnMapServer = make(map[string]net.UDPAddr)

var UdpConnMutex sync.RWMutex

func UdpForwardDisconnect(key string) {

	UdpConnMutex.Lock()
	tepUdp, ok := UdpConnMapClient[key]
	if !ok {
		tepUdp, ok = UdpConnMapServer[key]
	}

	delete(UdpConnMapClient, key)
	delete(UdpConnMapServer, key)

	delete(UdpConnMapClient, tepUdp.String())
	delete(UdpConnMapServer, tepUdp.String())
	UdpConnMutex.Unlock()

}

func UdpForwardHandler(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		fmt.Println("run udp server error:", err, "[", udpAddr, "]")
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("listen udp server error:", err)
		return
	}
	var buf [2000]byte

	for {
		msgLength, raddr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			continue
		}
		UdpConnMutex.RLock()
		//先查找Client的Map
		targetAddr, ok := UdpConnMapClient[raddr.String()]
		//如果没有找到
		if !ok {
			//没找到继续查找Server的
			targetAddr, ok = UdpConnMapServer[raddr.String()]
			//如果都没找到
			if !ok {
				UdpConnMutex.RUnlock()
				//可能是胡乱发的包
				if msgLength != 8 {
					continue
				}
				var userId uint64
				binary.Read(bytes.NewBuffer(buf[:msgLength]), binary.LittleEndian, &userId)
				currentUser, ok := UserMap[userId]
				//客户端bug
				if !ok {
					continue
				}
				currentUser.udpForwardAddr = *raddr
				UserMap[userId] = currentUser

				fmt.Println("User:", userId, "update the information")
				//这时候给他返回一个更新成功的Json
				forInfo := ForwardPortUpdate{time.Now().Unix()}
				strjson, _ := json.Marshal(forInfo)
				sendjson, _ := json.Marshal(JsonSend{SJForwardPortUpdate, string(strjson), SENoError})
				currentUser.websocketWriteChan <- string(sendjson)
				continue
			}
		}
		UdpConnMutex.RUnlock()
		fmt.Println("receive the:", raddr, "forward to ", (&targetAddr).String())
		//这里就进入转发阶段了
		conn.WriteToUDP(buf[0:msgLength], &targetAddr)
	}
}
