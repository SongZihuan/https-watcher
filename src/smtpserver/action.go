package smtpserver

import (
	"github.com/SongZihuan/https-watcher/src/logger"
)

func printError(err error) {
	if err == nil {
		return
	}

	logger.Errorf("SMTP Send Error: %s", err.Error())
}

func SendNotify(msg string) {
	printError(Send("HTTPS 到期通知", msg))
}
