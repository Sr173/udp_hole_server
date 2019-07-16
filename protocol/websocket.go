package protocol

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const Version = "1.3"

func websocketReader(conn *websocket.Conn, id uint64) {

	if _, ok := UserMap[id]; !ok {
		fmt.Println("currentUser don't exit")
		return
	}

	for {
		_, buf, err := conn.ReadMessage()
		if err != nil {
			UserMapMutex.Lock()
			close(UserMap[id].websocketWriteChan)
			currenctForwardPort, _ := UserMap[id]
			currenctForwardPortStr := (&currenctForwardPort.udpForwardAddr).String()
			tepUdp, ok := UdpConnMapClient[currenctForwardPortStr]
			if ok {
				delete(UdpConnMapClient, currenctForwardPortStr)
				delete(UdpConnMapServer, tepUdp.String())
			} else {
				tepUdp, ok := UdpConnMapServer[currenctForwardPortStr]
				if ok {
					delete(UdpConnMapServer, currenctForwardPortStr)
					delete(UdpConnMapClient, tepUdp.String())
				}
			}
			delete(UserMap, id)
			UserMapMutex.Unlock()
			fmt.Println("user is disconnect:", id, ", current user num:", len(UserMap))
			return
		}
		UserMapMutex.RLock()
		currentUser := UserMap[id]
		UserMapMutex.RUnlock()

		res, err := simplejson.NewJson(buf)
		if err != nil {
			fmt.Println(string(buf))
			fmt.Println("parse json error:", err)
			continue
		}

		switch dt, _ := res.Get("DataType").Int(); dt {
		case SJConnectRequest:
			if time.Now().Unix()-currentUser.Address[0].ConnTime > 5000 {
				connError := ServerError{SEAddressNeedUpdate, SJServerError}
				jsonStr, _ := json.Marshal(connError)
				currentUser.websocketWriteChan <- string(jsonStr)
				continue
			}

			targetUserId, err := res.Get("Data").Get("TargetUserId").Uint64()
			if err != nil {
				fmt.Println("get targetUserId id error:", err)
				continue
			}

			//先查询目标用户
			targetUser, ok := UserMap[targetUserId]
			if !ok {
				fmt.Println("target currentUser id is incorrect")
				continue
			}

			connReq := ConnectRequest{currentUser.Address[0], currentUser.UserId}
			tempConnJson, _ := json.Marshal(connReq)
			connJson, _ := json.Marshal(JsonSend{SJConnectRequest, string(tempConnJson), SENoError})
			targetUser.websocketWriteChan <- string(connJson)
			break
		case SJConnectReady:
			targetJsonString, err := res.Get("Data").String()

			if err != nil {
				fmt.Println("get data error:", err)
				continue
			}

			resJson, err := simplejson.NewJson([]byte(targetJsonString))

			if err != nil {
				fmt.Println("get data error:", err)
				continue
			}

			targetUserId, err := resJson.Get("TargetUserId").Uint64()

			if err != nil {
				fmt.Println("get targetUserId id error:", err)
				continue
			}

			fmt.Println("user ready to receive:", currentUser.UserId, "target user id:", targetUserId)

			//先查询目标用户
			targetUser, ok := UserMap[targetUserId]
			if !ok {
				fmt.Println("target currentUser id is incorrect")
				continue
			}

			targetUser.websocketWriteChan <- string(buf)
			break
		case SJFrowardInformation:
			forwardInfo := ForwardInformation{UdpForwardPort}
			tempConnJson, _ := json.Marshal(forwardInfo)
			connJson, _ := json.Marshal(JsonSend{SJFrowardInformation, string(tempConnJson), SENoError})
			currentUser.websocketWriteChan <- string(connJson)
			break
		case SJForwardRequest:
			targetUserId, err := res.Get("Data").Get("TargetUserId").Uint64()
			if err != nil {
				fmt.Println("get targetUserId id error:", err)
				continue
			}

			//先查询目标用户
			targetUser, ok := UserMap[targetUserId]
			if !ok {
				fmt.Println("target currentUser id is incorrect")
				continue
			}
			connReq := ForwardRequest{currentUser.UserId}
			tempReqJson, _ := json.Marshal(connReq)
			connJson, _ := json.Marshal(JsonSend{SJForwardRequest, string(tempReqJson), SENoError})
			targetUser.websocketWriteChan <- string(connJson)
			break
		case SJForwardConnectReady:
			targetJsonString, err := res.Get("Data").String()

			if err != nil {
				fmt.Println("get data error:", err)
				continue
			}

			resJson, err := simplejson.NewJson([]byte(targetJsonString))

			if err != nil {
				fmt.Println("get data error:", err)
				continue
			}

			targetUserId, err := resJson.Get("TargetUserId").Uint64()

			if err != nil {
				fmt.Println("get targetUserId id error:", err)
				continue
			}

			fmt.Println("user ready to receive:", currentUser.UserId, "target user id:", targetUserId)

			//先查询目标用户
			targetUser, ok := UserMap[targetUserId]
			if !ok {
				fmt.Println("target currentUser id is incorrect")
				continue
			}

			//这时候已经可以开始转发了
			UdpConnMapClient[currentUser.udpForwardAddr.String()] = targetUser.udpForwardAddr
			UdpConnMapClient[targetUser.udpForwardAddr.String()] = currentUser.udpForwardAddr

			targetUser.websocketWriteChan <- string(buf)
			break
		}

	}
}

var userId uint64 = 1
var userLocker sync.Mutex

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	c, err := u.Upgrade(w, r, nil)

	defer c.Close()

	if err != nil {
		return
	}

	userLocker.Lock()
	fmt.Println("get a new connect user id ", userId)
	currentUserId := userId
	userId++
	var user UserInformation
	user.ConnTime = time.Now().Unix()
	user.websocketWriteChan = make(chan string)
	user.UserId = currentUserId
	UserMapMutex.Lock()
	UserMap[currentUserId] = user
	UserMapMutex.Unlock()
	userLocker.Unlock()

	serverInfo := ServerInformation{Version, UdpPortList, currentUserId}
	strjson, _ := json.Marshal(serverInfo)
	sendjson, _ := json.Marshal(JsonSend{SJConnected, string(strjson), SENoError})
	c.WriteMessage(websocket.TextMessage, []byte(sendjson))

	go websocketReader(c, currentUserId)

	for {
		message, ok := <-user.websocketWriteChan
		if !ok {
			return
		}
		c.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
