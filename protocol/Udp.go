package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

var UdpPortList = [2]string{"1053", "1054"}

func UdpHandler(portIndex int) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+UdpPortList[portIndex])
	if err != nil {
		fmt.Println("run udp server error:", err, "[", udpAddr, "]")
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("listen udp server error:", err)
		return
	}
	var buf [200]byte
	for {
		msgLength, raddr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		if msgLength != 8 {
			fmt.Println("illegal format data")
			return
		}
		var userId uint64
		binary.Read(bytes.NewBuffer(buf[:msgLength]), binary.LittleEndian, &userId)
		user, ok := UserMap[userId]
		if ok {
			UserMapMutex.Lock()
			user.Address[portIndex].Port = raddr.Port
			user.Address[portIndex].Ip = raddr.IP
			user.Address[portIndex].ConnTime = time.Now().Unix()
			user.Address[portIndex].PortIndex = portIndex
			UserMap[userId] = user
			strjson, _ := json.Marshal(user.Address[portIndex])
			sendjson, _ := json.Marshal(JsonSend{SJPortUpdate, string(strjson), SENoError})
			user.websocketWriteChan <- string(sendjson)
			UserMapMutex.Unlock()
		} else {
			fmt.Print("received a incorrect user id")
		}
	}
}
