package httpswatcher

import (
	"errors"
	"github.com/SongZihuan/https-watcher/src/config"
	"github.com/SongZihuan/https-watcher/src/flagparser"
	"github.com/SongZihuan/https-watcher/src/logger"
	"github.com/SongZihuan/https-watcher/src/notify"
	"github.com/SongZihuan/https-watcher/src/utils"
	"github.com/SongZihuan/https-watcher/src/watcher"
	"os"
)

func MainV1() (exitcode int) {
	var err error

	err = flagparser.InitFlag()
	if errors.Is(err, flagparser.StopFlag) {
		return 0
	} else if err != nil {
		return utils.ExitByError(err)
	}

	if !flagparser.IsReady() {
		return utils.ExitByErrorMsg("flag parser unknown error")
	}

	cfgErr := config.InitConfig(flagparser.ConfigFile())
	if cfgErr != nil && cfgErr.IsError() {
		return utils.ExitByError(cfgErr)
	}

	if !config.IsReady() {
		return utils.ExitByErrorMsg("config parser unknown error")
	}

	err = logger.InitLogger(os.Stdout, os.Stderr)
	if err != nil {
		return utils.ExitByError(err)
	}

	if !logger.IsReady() {
		return utils.ExitByErrorMsg("logger unknown error")
	}

	err = notify.InitNotify()
	if err != nil {
		logger.Errorf("init notify fail: %s", err.Error())
		return 1
	}

	logger.Executablef("%s", "ready")
	logger.Infof("run mode: %s", config.GetConfig().GlobalConfig.GetRunMode())

	err = watcher.Run()
	if err != nil {
		logger.Errorf("run watcher fail: %s", err.Error())
		return 1
	}

	return 0
}
