package main

import (
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte("text" + "\n"))
}
