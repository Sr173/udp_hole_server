package main

import (
	"./protocol"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("server is start")
	go protocol.UdpHandler(0)
	go protocol.UdpHandler(1)
	go protocol.UdpForwardHandler(protocol.UdpForwardPort)

	http.HandleFunc("/ws/", protocol.WebsocketHandler)
	http.ListenAndServe(":8080", nil)
}
