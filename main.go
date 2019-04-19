package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var handshakeHeader = map[string]bool{
	"Host":                   true,
	"Upgrade":                true,
	"Connection":             true,
	"Sec-Websocket-Key":      true,
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

func main() {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		panic("dial error: " + err.Error())
	}
	fmt.Println(conn)
	br := bufio.NewReader(conn.(io.ReadWriteCloser))
	bw := bufio.NewWriter(conn.(io.ReadWriteCloser))
	res := handshake(br, bw)
	//client := &http.Client{Transport: defaultTransport}
	////opening handshake
	//p := make([]byte, 16)
	//_, _ = rand.Read(p)
	//req, _ := http.NewRequest("GET", url+"/socket.io/?transport=websocket", nil)
	//req.Header.Add("Upgrade", "websocket")
	//req.Header.Add("Origin", "http://localhost")
	//req.Header.Add("Connection", "Upgrade")
	//req.Header.Add("Sec-WebSocket-Key", base64.StdEncoding.EncodeToString(p))
	//req.Header.Add("Sec-WebSocket-Protocol", "chat")
	//req.Header.Add("Sec-WebSocket-Version", "13")
	//resp, err := client.Do(req)

	fmt.Println(res)
	//_, _ = websocket.NewConfig("ws://localhost:3000/socket.io/ws", "http://localhost:3000/")
	//conn, _ := net.Dial("tcp", "localhost:5000")
	//br := bufio.NewReader(conn)
	//bw := bufio.NewWriter(conn)
	//buf := bufio.NewReadWriter(br, bw)
	//ws := newHybiConn(config, buf, conn, nil)
	fmt.Println(conn.LocalAddr())

	go func() {
		for {
			readFrame(conn)
		}
	}()
	frame := createFrame("hullo")
	//time.Sleep(3 * time.Second)
	_, err = conn.Write(frame)
	if err != nil {
		fmt.Println(err.Error())
	}
	//print("hullo")
	//readFrame(conn)
	time.Sleep(1 * time.Second)
	frame = createFrame("hallo")
	_, err = conn.Write(frame)
	if err != nil {
		fmt.Println(err.Error())
	}
	readFrame(conn)
	//sendMessage := make(chan string)
	ch := make(chan string)

	//go func() {
	//	for {
	//		nyan := <-sendMessage
	//		conn.Write(createFrame(nyan))
	//	}
	//}()
	<-ch
}

func readFrame(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if n == 0 {
		return
	}
	if err != nil {
		fmt.Printf("error: " + err.Error())
	}
	fmt.Printf("n = %d\n", n)
	fmt.Printf("FIN: %d\n", refbit(buf[0], 7))
	fmt.Printf("opcode: %x\n", buf[0]&0x0F)
	if buf[0]&0x0F == 0x8 {
		fmt.Println(buf[2:])
		fmt.Println("close")
		_, err := conn.Write(createCloseFrame())
		if err != nil {
			fmt.Println("error: " + err.Error())
		}
	} else {
		fmt.Println("test")
		//ch <- "test"
	}
	fmt.Printf("MASK: %b\n", buf[1]&0x80)
	fmt.Printf("Payload len: %b\n", buf[1]&0x7F)
	fmt.Printf("Payload len: %d\n", buf[1]&0x7F)
	fmt.Printf("Payload Data: {%s\n\n}", string(buf[:]))
}

func createCloseFrame() []byte {
	message := []byte{0x08}
	frame := make([]byte, 6)
	frame[0] = 0x80 | 0x08
	mask := byte(0x80)
	frame[1] = mask | 0x02
	frame[2] = 0xAA
	frame[3] = 0xBB
	frame[4] = 0xCC
	frame[5] = 0xDD
	for i := 0; i < len(message); i++ {
		message[i] = message[i] ^ frame[2+i%4]
	}
	frame = append(frame, message...)
	return frame
}

func createFrame(str string) []byte {
	message := []byte(str)
	frame := make([]byte, 6)
	frame[0] = 0x80 | 0x01
	frame[1] = 0x80 | byte(len(message))
	frame[2] = 0xAA
	frame[3] = 0xBB
	frame[4] = 0xCC
	frame[5] = 0xDD
	for i := 0; i < len(message); i++ {
		message[i] = message[i] ^ frame[2+i%4]
	}
	frame = append(frame, message...)
	return frame
}

func refbit(i byte, b uint) byte {
	return (i >> b) & 1
}

func handshake(br *bufio.Reader, bw *bufio.Writer) (res *http.Response) {
	p := make([]byte, 16)
	_, _ = rand.Read(p)
	//	bw.WriteString("GET /socket.io/?transport=websocket HTTP/1.1\r\n")
	bw.WriteString("GET /echo HTTP/1.1\r\n")
	bw.WriteString("Host: localhost\r\n")
	bw.WriteString("Upgrade: websocket\r\n")
	bw.WriteString("Connection: Upgrade\r\n")
	bw.WriteString("Origin: http://localhost\r\n")
	bw.WriteString("Sec-WebSocket-Protocol: chat\r\n")
	bw.WriteString("Sec-WebSocket-Version: 13\r\n")
	bw.WriteString("Sec-WebSocket-Key: " + base64.StdEncoding.EncodeToString(p) + "\r\n")
	header := &http.Header{}
	_ = header.WriteSubset(bw, handshakeHeader)
	bw.WriteString("\r\n")
	if err := bw.Flush(); err != nil {
		fmt.Println(err)
	}
	res, _ = http.ReadResponse(br, &http.Request{Method: "GET"})
	return
}
