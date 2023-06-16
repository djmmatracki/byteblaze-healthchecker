package app

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/djmmatracki/byteblaze-healtchecker/dockerclient"
	"github.com/djmmatracki/byteblaze-healtchecker/gitclient"
	"github.com/djmmatracki/byteblaze-healtchecker/healthcheck"
	"github.com/jackpal/bencode-go"
	"github.com/sirupsen/logrus"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
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

func Run(logger *logrus.Logger, ipAddress string, port int) {
	address := fmt.Sprintf("%s:%d", ipAddress, port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("error while starting to listen on")
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Errorf("error while handeling connection %v", err)
		}
		logger.Debugln("received connection")
		go handleConnection(logger, conn)
	}
}

func handleConnection(logger *logrus.Logger, conn net.Conn) {
	defer conn.Close()
	// Read infohash
	infoHash, err := io.ReadAll(conn)
	if err != nil {
		logger.Errorf("error while reading infohash: %v", err)
		return
	}
	if len(infoHash) != 20 {
		logger.Errorf("infohash not expeceted length: %d", len(infoHash))
		return
	}

	// Load tf file
	f, err := os.Open(
		fmt.Sprintf("/var/byteblaze/%x/torrentfile", infoHash),
	)
	if err != nil {
		logger.Errorf("cannot open torrentfile: %v", err)
		return
	}
	defer f.Close()

	tf := TorrentFile{}
	err = bencode.Unmarshal(f, &tf)
	if err != nil {
		logger.Errorf("cannot unmarshal torrentfile: %v", err)
		return
	}
	if ok := healthcheck.ValidateSupportedApps(tf.App); !ok {
		// App is not supported
		logger.Errorf("app not supported: %s", tf.App)
		return
	}
	// Create dir and repo if doesnt exist
	destinationDirectory := path.Join("/etc/byteblaze", tf.App)
	repo, err := gitclient.PrepareDestinationDirectory(destinationDirectory)
	if err != nil {
		logger.Errorf("error while preparing directory: %v", err)
		return
	}

	err = replaceFile(&tf, infoHash)
	if err != nil {
		logger.Errorf("error while replacing file: %v", err)
		return
	}

	// Restart docker container
	err = dockerclient.RestartConatiners(tf.App)
	if err != nil {
		logger.Errorf("error while restarting containers: %v", err)
		return
	}
	time.Sleep(5 * time.Second)

	// Healthcheck
	err = healthcheck.ApplyHealthchecks(tf.App)
	if err != nil {
		// Rollback changes
		err = gitclient.RollbackChanges(repo)
		if err != nil {
			// Add retry
			logger.Errorf("error while rolling back changes: %v", err)
			return
		}

		err = dockerclient.RestartConatiners(tf.App)
		if err != nil {
			logger.Errorf("error while restarting containers: %v", err)
			return
		}
		return
	}
	err = gitclient.CommitAllChanges(repo)
	if err != nil {
		logger.Errorf("commit all changes: %v", err)
		return
	}
}

func replaceFile(tf *TorrentFile, infoHash []byte) error {
	dropPath := strings.TrimPrefix(tf.DropPath, fmt.Sprintf("/etc/%s", tf.App))
	sourceFile := fmt.Sprintf("/var/byteblaze/%x/%s", infoHash, tf.Info.Name)
	destinationFile := path.Join("/etc/byteblaze", tf.App, dropPath)

	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}
	return nil
}
