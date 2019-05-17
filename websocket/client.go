package websocket

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
)

type WSClient struct {
	conn *net.Conn
}

func (self *WSClient) Conn(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic("dial error: " + err.Error())
	}
	fmt.Println(conn)
	br := bufio.NewReader(conn.(io.ReadWriteCloser))
	bw := bufio.NewWriter(conn.(io.ReadWriteCloser))
	res := handshake(br, bw)
	fmt.Println(res)
	fmt.Println(conn.LocalAddr())
	self.conn = &conn
}

func (self *WSClient) Send(frame []byte) {
	_, err := (*self.conn).Write(frame)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (self *WSClient) Recieve() {
	readFrame(*self.conn)
}
func handshake(br *bufio.Reader, bw *bufio.Writer) (res *http.Response) {
	p := make([]byte, 16)
	_, _ = rand.Read(p)
	//	bw.WriteString("GET /socket.io/?transport=websocket HTTP/1.1\r\n")
	bw.WriteString("GET /echo HTTP/1.1\r\n")
	header := &http.Header{}
	header.Set("Host", "localhost")
	header.Set("Upgrade", "websocket")
	header.Set("Connection", "Upgrade")
	header.Set("Origin", "http://localhost")
	header.Set("Sec-WebSocket-Protocol", "chat")
	header.Set("Sec-WebSocket-Version", "13")
	header.Set("Sec-WebSocket-Key", base64.StdEncoding.EncodeToString(p))
	_ = header.Write(bw)
	bw.WriteString("\r\n")
	// ラップしているio.Writerへバッファリングしているデータを書き込む
	if err := bw.Flush(); err != nil {
		fmt.Println(err)
	}
	res, _ = http.ReadResponse(br, &http.Request{Method: "GET"})
	return
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
func refbit(i byte, b uint) byte {
	return (i >> b) & 1
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
func CreateFrame(str string) []byte {
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
