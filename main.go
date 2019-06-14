package main

import (
	"ws/websocket"
	"net"
	"net/http"
	"time"
)

var handshakeHeader = map[string]bool{
	"Host":                   true,
	"Upgrade":                true,
	"Connection":             true,
	"Sec-Websocket-Origin":   true,
	"Sec-Websocket-Version":  true,
	"Sec-Websocket-Protocol": true,
	"Sec-Websocket-Accept":   true,
}
var port = "12345"
var url = "http://localhost" + port
var addr, _ = net.ResolveTCPAddr("tcp", "localhost:5000")
var defaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		LocalAddr: addr,
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// websocket client
func main() {
	client := &websocket.WSClient{}
	client.Conn("localhost:" + port)

	go func() {
		for {
			client.Recieve()
		}
	}()
	frame := websocket.CreateFrame("hullo")
	client.Send(frame)
	time.Sleep(1 * time.Second)
	frame = websocket.CreateFrame("hallo")
	client.Send(frame)
	ch := make(chan string)

	<-ch
}
