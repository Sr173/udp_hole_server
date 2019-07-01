package main

import (
	"fmt"
	"net"
	"time"
)

type peerInfomation struct {
	lastUpdateTime int64
	conn           net.UDPAddr
	isHole         bool
}

var clientMap map[string]peerInfomation

func dealWithUdpGetMsg(conn *net.UDPConn, udpAddr *net.UDPAddr, msg string) {
	user := msg[1:]
	if msg[0] == '0' {
		fmt.Println("用户更新了信息:", user, ":", udpAddr.String())
		var temp peerInfomation
		temp.conn = *udpAddr
		temp.lastUpdateTime = time.Now().Unix()
		clientMap[user] = temp
		_, _ = conn.WriteToUDP([]byte("0"+udpAddr.String()), udpAddr)
	} else {
		userAddr, ok := clientMap[user]
		if !ok {
			conn.WriteToUDP([]byte("-"), udpAddr)
			fmt.Println("获取用户信息失败")
			return
		}
		if msg[0] == '1' {
			conn.WriteToUDP([]byte("1"+userAddr.conn.String()), udpAddr)
		} else if msg[0] == '2' { //想要打洞
			print("接收到用户[", udpAddr.String(), "]的打洞请求，打洞目标", userAddr.conn.String())
			conn.WriteToUDP([]byte("2"+udpAddr.String()), &userAddr.conn)
		}
		clientMap[user] = userAddr
	}
}

func main() {

	clientMap = make(map[string]peerInfomation)

	udpAddr, err := net.ResolveUDPAddr("udp", ":543")

	if err != nil {
		fmt.Print("bind error")
	}

	conn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		fmt.Print(err)
		return
	}

	var buf [200]byte
	for {
		msgLength, raddr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}

		go dealWithUdpGetMsg(conn, raddr, string(buf[0:msgLength]))
	}

}
