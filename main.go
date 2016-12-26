package main

import "net"
import "bytes"

func main() {
	b := bytes.NewBufferString("hello")
	buf := make([]byte, 2)
	b.Read(buf)
	b.Read(buf)

	conn, err := net.Dial("tcp", "192.168.102.73:8081")
	if err != nil {
		// handle error
	}

	conn.Write([]byte("Hello"))
}
