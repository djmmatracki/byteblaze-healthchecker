package main

import (
	"github.com/djmmatracki/byteblaze-healtchecker/app"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	app.Run(logger, "0.0.0.0", 6881)
}
