package protocol

import (
	"net"
	"sync"
)

type UserUdpInformation struct {
	Port      int
	ConnTime  int64
	Ip        net.IP
	PortIndex int
}

type UserInformation struct {
	ConnTime           int64
	Address            [2]UserUdpInformation
	websocketWriteChan chan string
	UserId             uint64
	udpForwardAddr     net.UDPAddr
}

type ServerInformation struct {
	Version string
	UdpPort [2]string
	UserId  uint64
}

type OtherConnectRequest struct {
	TargetUserId uint64
}

type JsonSend struct {
	DataType  ServerJsonHeader
	Data      string
	ErrorCode ServerErrorCode
}

type ServerError struct {
	ErrorCode  ServerErrorCode
	RecvHeader ServerJsonHeader
}

type ConnectRequest struct {
	Requester   UserUdpInformation
	RequesterId uint64
}

type ForwardInformation struct {
	UdpPort string
}

type ForwardRequest struct {
	TargetUserId uint64
}

type ForwardPortUpdate struct {
	ConnTime int64
}

type ServerJsonHeader int

const (
	SJConnected = iota
	SJConnectRequest
	SJPortUpdate
	SJServerError
	SJConnectReady
	SJFrowardInformation
	SJForwardRequest
	SJForwardPortUpdate
	SJForwardConnectReady
)

type ServerErrorCode int

const (
	SENoError = iota
	SEUserNotExit
	SEAddressNeedUpdate
)

var UserMap = make(map[uint64]UserInformation)
var UserMapMutex sync.RWMutex
