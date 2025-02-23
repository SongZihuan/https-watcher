package wxrobot

import (
	"github.com/SongZihuan/https-watcher/src/logger"
)

func printError(err error) {
	if err == nil {
		return
	}

	logger.Errorf("WxRobot Send Error: %s", err.Error())
}

func SendNotify(msg string) {
	printError(Send(msg, true))
}
