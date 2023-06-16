package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	bencode "github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int64  `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type TorrentFile struct {
	Announce string      `bencode:"announce"`
	Version  string      `bencode:"version"`
	DropPath string      `bencode:"drop_path"`
	App      string      `bencode:"app"`
	Info     bencodeInfo `bencode:"info"`
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func main() {
	data, err := ioutil.ReadFile("default.conf")
	if err != nil {
		log.Fatal(err)
	}

	pieceLength := int64(256 * 1024) // typically piece sizes are powers of two, like 256KB
	numPieces := len(data) / int(pieceLength)
	if len(data)%int(pieceLength) != 0 {
		numPieces++
	}

	pieces := ""
	for i := 0; i < numPieces; i++ {
		start := i * int(pieceLength)
		end := start + int(pieceLength)
		if end > len(data) {
			end = len(data)
		}
		hash := sha1.Sum(data[start:end])
		pieces += string(hash[:])
	}

	info := bencodeInfo{
		PieceLength: pieceLength,
		Pieces:      pieces,
		Length:      len(data),
		Name:        "default.conf",
	}

	torrent := TorrentFile{
		Announce: "http://172.104.234.48:6969/announce",
		DropPath: "/etc/nginx/conf.d/default.conf",
		Version:  "0.0.1",
		App:      "nginx",
		Info:     info,
	}
	h, err := info.hash()
	fmt.Printf("%x\n", h)

	file, err := os.Create("torrentfile")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = bencode.Marshal(file, torrent)
	if err != nil {
		log.Fatal(err)
	}
}
