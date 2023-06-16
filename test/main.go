package main

import (
	"fmt"
	"net"

	"github.com/djmmatracki/byteblaze-healtchecker/test/torrent"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6881")
	if err != nil {
		fmt.Printf("error while dialing %v\n", err)
		return
	}
	tf, err := torrent.Open("torrentfile")
	if err != nil {
		fmt.Printf("error while opening %v\n", err)
		return
	}

	fmt.Printf("sending info hash with length %x", tf.InfoHash)
	_, err = conn.Write(tf.InfoHash[:])
	if err != nil {
		return
	}
	conn.Close()
}
