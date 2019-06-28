package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type peerInfomation struct{
	lastUpdateTime int64
	conn net.UDPAddr
};

var clientMap map[string]peerInfomation


var mutex sync.Mutex

func dealWithUdpGetMsg(conn *net.UDPConn, udpAddr *net.UDPAddr,msg string){
	mutex.Lock()
	defer mutex.Unlock()
	user := msg[1:]
	if msg[0] == '0' {
		fmt.Println("用户更新了信息:",user,":",udpAddr.String())
		var temp peerInfomation
		temp.conn = *udpAddr
		temp.lastUpdateTime = time.Now().Unix()
		clientMap[user] = temp;

		_, _ = conn.WriteToUDP([]byte("0" + udpAddr.String()), udpAddr)
	} else {
		userAddr,ok := clientMap[user]
		if !ok{
			conn.WriteToUDP([]byte("-"), udpAddr)
			fmt.Println("获取用户信息失败")

			return
		}
		if msg[0] == '1' {
			conn.WriteToUDP([]byte("1" + userAddr.conn.String()), udpAddr)
		} else if msg[0] == '2' { //想要打洞
			print("接收到用户[",udpAddr,"]的打洞请求，打洞目标",&userAddr.conn)
			conn.WriteToUDP([]byte("2" + udpAddr.String()), &userAddr.conn)
		}
	}
}

func heartBeat(conn *net.UDPConn){
	for {
		mutex.Lock()
		fmt.Println("发送心跳包")
		for k, v := range (clientMap) {
			if (time.Now().Unix()-v.lastUpdateTime > 5000) {
				fmt.Println("删除不活跃用户:", k, "-", v.conn)
				delete(clientMap, k)
			} else {
				conn.WriteToUDP([]byte("H"), &v.conn)
			}
		}
		mutex.Unlock()
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func main() {

	clientMap = make(map[string]peerInfomation)

	udpAddr, err := net.ResolveUDPAddr("udp", ":543")

	udpAddr1, err := net.ResolveUDPAddr("udp", ":544")

	if err != nil {
		fmt.Print("bind error")
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	conn2, err := net.ListenUDP("udp", udpAddr1)

	if err != nil {
		fmt.Print(err)
		return
	}

	go heartBeat(conn)

	go  func (udpConn *net.UDPConn) {
		var buf [200]byte
		for {
			msgLength, raddr, err := conn.ReadFromUDP(buf[0:])
			if err != nil {
				fmt.Println(err)
				return
			}

			go dealWithUdpGetMsg(conn, raddr, string(buf[0:msgLength]))
		}
	}(conn2)


	var buf [200]byte
	for {
		msgLength, raddr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}

		go dealWithUdpGetMsg(conn,raddr,string(buf[0:msgLength]))
	}


}